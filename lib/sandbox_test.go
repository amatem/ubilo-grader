package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestAddFile(t *testing.T) {
	sandbox, err := NewSandbox()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer sandbox.Remove()

	testfile, err := os.Open(filepath.Join(config.BasePath, "sol.cpp"))
	if err != nil {
		t.Errorf("internal error: %+v\n", err)
		t.FailNow()
	}

	err = sandbox.AddFile(testfile, "test.cpp", false)
	if err != nil {
		t.Errorf("addFile Failed: %+v\n", err)
		t.FailNow()
	}

	testget, err := sandbox.GetFile("test.cpp")
	if err != nil {
		t.Errorf("getFile Failed: %+v\n", err)
		t.FailNow()
	}

	_, err = ioutil.ReadAll(testget)
	if err != nil {
		t.Errorf("readall Failed: %+v\n", err)
		t.FailNow()
	}
}

func TestCompileAndRun(t *testing.T) {
	sandbox, err := NewSandbox()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer sandbox.Remove()
	log.Println("id: ", sandbox.box_id)

	addFile := func(name string, isExec bool, path ...string) {
		paths := []string{config.BasePath}
		paths = append(paths, path...)
		file, err := os.Open(filepath.Join(paths...))
		if err != nil {
			t.Errorf("internal error: %+v\n", err)
			t.FailNow()
		}
		err = sandbox.AddFile(file, name, isExec)
		if err != nil {
			t.Errorf("adding file to sandbox failed: %+v\n", err)
			t.FailNow()
		}
	}

	addFile("sol", true, "a.out")
	addFile("input.txt", false, "batch", "io", "kalmax.1.gir")

	infile := "input.txt"
	res, err := sandbox.RunCommand([]string{"./sol"}, &infile, nil)
	if err != nil {
		t.Errorf("runcommand error: %+v\n", err)
		t.FailNow()
	}

	fmt.Println(res)
}
