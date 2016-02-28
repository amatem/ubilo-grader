package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ERR_ACCEPT = iota
	ERR_TLE
	ERR_SEGV
	ERR_INT
)

type ErrorCode uint8

type Sandbox interface {
	AddFile(file *os.File, name string, isExecutable bool) error
	SetTime(time int64) error
	SetMemory(memory int64) error
	RunCommand(execargs []string, infile, outfile *string) (Stats, error)
	GetFile(name string) (*os.File, error)
	Clean() error
	Remove() error
	GetTempDir() string
}

type Stats struct {
	status ErrorCode
	memory int
	time   int64
}

type sandbox struct {
	exec_path    string
	box_path     string
	box_id       uint
	time_limit   *int64
	memory_limit *int64
	temp_path    string
}

var boxIds int64

func NewSandbox() (*sandbox, error) {
	boxID := -1

	for i := 0; i < 10; i++ {
		if (boxIds & (1 << uint(i))) == 0 {
			boxID = i
			break
		}
	}

	if boxID == -1 {
		return nil, fmt.Errorf("No available box id is found")
	}

	time_def_limit := int64(1000)
	memory_def_limit := int64(256)
	sb := &sandbox{
		time_limit:   &time_def_limit,
		memory_limit: &memory_def_limit,
	}
	sb.box_id = uint(boxID)
	sb.exec_path = filepath.Join(config.BasePath, "jail", "isolate")
	out, err := exec.Command(sb.exec_path, fmt.Sprintf("--box-id=%d", sb.box_id), "--init").Output()
	if err != nil {
		log.Println("Sandbox init output: ", out)
		return nil, err
	}
	boxIds |= 1 << uint(boxID)
	sb.box_path = filepath.Join("/tmp/box", strconv.Itoa(int(sb.box_id)), "box")
	sb.temp_path, err = ioutil.TempDir(filepath.Join(config.BasePath, "jail", "temp"), "jail")
	if err != nil {
		return nil, err
	}
	return sb, nil
}

func (s *sandbox) AddFile(file *os.File, name string, isExecutable bool) error {
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if !stat.Mode().IsRegular() {
		return fmt.Errorf("err: is not a regular file")
	}

	dst_path := filepath.Join(s.box_path, name)
	dst, err := os.Create(dst_path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return err
	}

	if isExecutable {
		out, err := exec.Command("chmod", "+x", dst_path).Output()
		if err != nil {
			log.Println("chmod Error: ", out)
			return err
		}
	}

	return nil
}

func (s *sandbox) GetTempDir() string {
	return s.temp_path
}

func (s *sandbox) GetFile(name string) (*os.File, error) {
	path := filepath.Join(s.box_path, name)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s *sandbox) SetTime(time int64) error {
	s.time_limit = &time
	return nil
}

func (s *sandbox) SetMemory(memory int64) error {
	s.memory_limit = &memory
	return nil
}

func (s *sandbox) RunCommand(execargs []string, infile, outfile *string) (*Stats, error) {
	args := []string{
		"-t",
		fmt.Sprintf("%f", float64(*s.time_limit)/float64(1000)),
		"-m",
		fmt.Sprintf("%d", *s.memory_limit*1024),
	}

	if infile != nil {
		args = append(args, []string{
			"-i",
			*infile,
		}...)
	}
	if outfile != nil {
		args = append(args, []string{
			"-i",
			*outfile,
		}...)
	}
	//args = append(args, fmt.Sprintf("--chdir=%s", filepath.Join(config.BasePath, "jail")))
	//GUZEL BI BOXLAMA YAPISI LAZIM TEMP FOLDER FALAN FILAN
	//FIX ME
	args = append(args, fmt.Sprintf("--meta=%s", filepath.Join(s.temp_path, "log.txt")))
	args = append(args, fmt.Sprintf("--box-id=%d", s.box_id))
	args = append(args, "--run")
	args = append(args, "--")
	if execargs != nil {
		args = append(args, execargs...)
	}

	fmt.Println(args)
	cmd := exec.Command(s.exec_path, args...)
	_, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	logreader, err := os.Open(filepath.Join(s.temp_path, "log.txt"))
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(logreader)
	if err != nil {
		return nil, err
	}

	result := &Stats{}
	statar := strings.Split(string(data), "\n")
	for _, stat := range statar {
		pair := strings.Split(stat, ":")
		if pair[0] == "status" {
			switch pair[1] {
			case "RE":
				result.status = ERR_SEGV
			case "SG":
				result.status = ERR_SEGV
			case "TO":
				result.status = ERR_TLE
			case "XX":
				result.status = ERR_INT
			}
		} else if pair[0] == "max-rss" {
			result.memory, err = strconv.Atoi(pair[1])
			if err != nil {
				return nil, err
			}
			result.memory /= 1024
		} else if pair[0] == "time" {
			sec, err := strconv.ParseFloat(pair[1], 64)
			if err != nil {
				return nil, err
			}
			result.time = int64(sec * 1000)
		}
	}

	return result, nil
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *sandbox) Clean() error {
	return removeContents(s.box_path)
}

func (s *sandbox) Remove() error {
	os.RemoveAll(s.temp_path)
	args := []string{
		fmt.Sprintf("--box-id=%d", s.box_id),
		"--cleanup",
	}

	_, err := exec.Command(s.exec_path, args...).Output()
	if err != nil {
		return err
	}

	boxIds ^= 1 << s.box_id
	return nil
}
