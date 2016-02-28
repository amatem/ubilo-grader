package lib

import "os"

type Storage interface {
	GetSubmission(subID string) (*os.File, error)
	GetGraderFiles(taskID string) ([]*os.File, error)
	GetInput(taskID string, num int64) (*os.File, error)
	GetOutput(taskID string, num int64) (*os.File, error)
	GetHelpers(taskID string) ([]*os.File, error)
	GetCompileCommand(taskID string) (string, error)
	GetGradeCommand(taskID string) (string, error)
	//SaveResult(subID string) error
}

type storage struct {
}
