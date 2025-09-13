package environment

import (
	"fmt"
	"os"
	"path/filepath"
)

type Env interface {
	CurrentDir() (string, error)
	ChangeDir(path string) error
	HomeDir() (string, error)
	Variable(key string) string
}

type SystemEnv struct{}

func NewSystemEnv() *SystemEnv {
	return &SystemEnv{}
}

func (e *SystemEnv) CurrentDir() (string, error) {
	return os.Getwd()
}

func (e *SystemEnv) ChangeDir(path string) error {
	return os.Chdir(path)
}

func (e *SystemEnv) HomeDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME not set")
	}
	return home, nil
}

func (e *SystemEnv) Variable(key string) string {
	return os.Getenv(key)
}

func (e *SystemEnv) DirName() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Base(wd), nil
}
