package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

/*func cp(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}*/
func cp(src, dst string) (int64, error) {
	src_file, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer src_file.Close()

	src_file_stat, err := src_file.Stat()
	if err != nil {
		return 0, err
	}

	if !src_file_stat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	dst_file, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dst_file.Close()
	return io.Copy(dst_file, src_file)
}

func main() {
	basePath, err := filepath.Abs("/home/reink/go/src/github.com/amatem/ubilo-grader/testgrader/warzone")
	if err != nil {
		log.Fatalln("basePath cannot be reached")
	}

	cmd := exec.Command("g++", "sol.cpp")
	cmd.Dir = basePath
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalln(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalln("compile error: ", err)
	}
	outb, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatalln(err)
	}
	errb, err := ioutil.ReadAll(stderr)
	if err != nil {
		log.Fatalln(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalln("wait error: ", err)
	}

	fmt.Println(outb)
	fmt.Println(errb)
	code := "kalmax"

	out, err := exec.Command(filepath.Join(basePath, "jail", "isolate", "isolate"), "--init").Output()
	defer func() {
		fmt.Println("yarrak")
	}()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("basarili ", out)

	boxpath := "/tmp/box/0/box"
	execfile := filepath.Join(basePath, "a.out")
	_, err = cp(execfile, filepath.Join(boxpath, "sol"))
	if err != nil {
		log.Fatalln("cp err: ", err)
	}
	output, err := exec.Command("chmod", "+x", filepath.Join(boxpath, "sol")).Output()
	if err != nil {
		log.Fatalln("chmod err: ", err)
	}
	fmt.Println(output)

	for i := 1; i <= 1; i++ {
		infile := filepath.Join(basePath, "batch", "io", code+"."+strconv.Itoa(i)+".gir")
		outfile := filepath.Join(basePath, "batch", "io", code+"."+strconv.Itoa(i)+".cik")
		_, err = cp(infile, filepath.Join(boxpath, "input.txt"))
		if err != nil {
			log.Fatalln("cp err: ", err)
		}
		args := []string{
			"--processes=0",
			"--full-env",
			"-i",
			"input.txt",
			"-o",
			"output.txt",
			"--run",
			"--",
			"./sol",
		}
		cmd := exec.Command(filepath.Join(basePath, "jail", "isolate", "isolate"), args...)
		fmt.Println(cmd.Args)
		out, err := cmd.Output()
		fmt.Println(out)
		if err != nil {
			log.Fatalln("iso run error: ", err)
		}
		cp(filepath.Join(boxpath, "output.txt"), filepath.Join(basePath, "jail", "output.txt"))
		outfile = outfile
	}
}
