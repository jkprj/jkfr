package utils

import (
	"encoding/json"
	"errors"

	jklog "github.com/jkprj/jkfr/log"
	jkos "github.com/jkprj/jkfr/os"

	"github.com/BurntSushi/toml"
)

// 读取并解析配置文件
func ReadConfigFile(path string, cfg interface{}) (err error) {

	if "" == path {
		jklog.Debugw("config file path is empty")
		return errors.New("config file path is empty")
	}

	buff, err := jkos.ReadFile(path)
	if nil != err {
		jklog.Errorw("jkos.ReadFile fail", "ConfigPath", path, "err", err)
		return err
	}

	var jsErr error
	var tmErr error

	defer func() {
		if nil == jsErr || nil == tmErr {
			// jklog.Infow("load config success", "cfg", cfg, "path", path)
		} else {
			jklog.Errorw("load config fail", "ConfigPath", path, "json_err", jsErr, "toml_err", tmErr)
		}
	}()

	jsErr = json.Unmarshal(buff, cfg)
	if nil == jsErr {
		return jsErr
	}

	_, tmErr = toml.Decode(string(buff), cfg)
	if nil != tmErr {
		jklog.Warnw("toml.Decode fail", "ConfigPath:", path, "err:", err)
		return tmErr
	}

	return nil
}
