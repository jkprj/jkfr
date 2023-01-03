package os

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func CurPath() string {
	path, _ := filepath.Abs(os.Args[0])
	return path
}

func CurDir() string {
	return filepath.Dir(CurPath())
}

func AppName() string {
	return filepath.Base(CurPath())
}

func ReadFile(path string) ([]byte, error) {
	var err error
	path, err = filepath.Abs(path)
	if nil != err {
		return nil, err
	}

	buff, err := ioutil.ReadFile(path)
	if nil != err {
		return nil, err
	}

	return buff, nil
}

func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
