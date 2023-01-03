package log

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/golang/glog"
)

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func getCurrentExePath() (string, error) {
	file, err := exec.LookPath(os.Args[0])

	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New("error: Can't find '/' or '\\'")
	}
	return string(path[0 : i+1]), nil
}
func getDefaultLogPath() (string, error) {
	curPath, err := getCurrentExePath()
	if err != nil {
		return "", err
	}
	logPath := curPath + "logs"
	isExist, err := pathExists(logPath)
	if err != nil {
		return "", err
	}
	if !isExist {
		//mkdir
		if err := os.Mkdir(logPath, os.ModePerm); err != nil {
			return "", err
		}
	}
	return logPath, nil
}
func getExeBaseName() (string, error) {
	var exeName string
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	base := filepath.Base(path)
	sysType := runtime.GOOS
	if sysType == "windows" {
		exeName = base[0 : len(base)-4]
	}
	if sysType == "linux" {
		exeName = base
	}
	return exeName, nil
}

//检查配置文件是否存在
func checkConfigFile(cfgOption *ZapOptions) bool {
	if _, err := getDefaultLogPath(); err != nil {
		return false
	}
	curPath, err := getCurrentExePath()
	if err != nil {
		return false
	}
	confPath := curPath + "conf/"
	isExist, err := pathExists(confPath)
	if err != nil || !isExist {
		return false
	}
	baseName, err := getExeBaseName()
	if err != nil {
		return false
	}
	cfgFile := confPath + baseName + ".json"
	isExist, err = pathExists(cfgFile)
	if err != nil || !isExist {
		return false
	}
	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		glog.Errorf("read conf file failed, err: %s, conf file: %s", err.Error(), cfgFile)
		return false
	}
	if err := json.Unmarshal(data, cfgOption); err != nil {
		glog.Errorf("unmarshal conf json data to struct failed, err: %s", err.Error())
		return false
	}
	return true
}
