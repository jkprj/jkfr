package log

import (
	kitlog "jkfr/gokit/log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
var core zapcore.Core

type ZapOptions struct {
	LogPath   string `json:"LogPath"`
	FileName  string `json:"FileName"`
	MaxSize   int    `json:"MaxSize"`
	MaxAge    int    `json:"MaxAge"`
	MaxBackup int    `json:"MaxBackup"`
	LogLevel  string `json:"LogLevel"`
}

var LogLevel = map[string]zapcore.Level{"info": zap.InfoLevel, "debug": zap.DebugLevel, "warning": zap.WarnLevel, "error": zap.ErrorLevel}

func init() {
	InitLogger()
}

func checkInitOptions(ops *ZapOptions) error {
	var err error
	if ops.LogPath == "" {
		ops.LogPath, err = getDefaultLogPath()
		if err != nil {
			return err
		}
	}
	if ops.FileName == "" {
		ops.FileName, err = getExeBaseName()
		if err != nil {
			return err
		}
	}
	if ops.MaxAge <= 0 {
		ops.MaxAge = 7
	}
	if ops.MaxBackup <= 0 {
		ops.MaxBackup = 5
	}
	if ops.MaxSize <= 0 {
		ops.MaxSize = 100
	}
	return nil
}

func initCfgOptions(args ...*ZapOptions) (*ZapOptions, error) {
	var err error
	ops := &ZapOptions{}
	//检查配置文件是否存在，使用配置文件配置
	if checkConfigFile(ops) {
	} else if len(args) == 0 { //使用默认配置
		if ops.LogPath, err = getDefaultLogPath(); err != nil {
			return nil, err
		}
		if ops.FileName, err = getExeBaseName(); err != nil {
			return nil, err
		}
		ops.MaxSize = 100
		ops.MaxBackup = 5
		ops.MaxAge = 7

	} else if len(args) == 1 { //
		ops = args[0]
	}
	//对参数合法性进行检查
	checkInitOptions(ops)
	return ops, nil
}

func InitLogger(args ...*ZapOptions) error {
	var ops *ZapOptions
	var err error
	if ops, err = initCfgOptions(args...); err != nil {
		return err
	}
	initZapLog(ops)
	kitlog.InitLogger(logger)
	return nil
}

func Debug(args ...interface{}) {
	logger.Sugar().Debug(args)
}

func Debugf(template string, args ...interface{}) {
	logger.Sugar().Debugf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Sugar().Debugw(msg, keysAndValues...)
}

func Info(args ...interface{}) {
	logger.Sugar().Info(args)
}

func Infof(template string, args ...interface{}) {
	logger.Sugar().Infof(template, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	logger.Sugar().Infow(msg, keysAndValues...)
}
func Warn(args ...interface{}) {
	logger.Sugar().Warn(args)
}

func Warnf(template string, args ...interface{}) {
	logger.Sugar().Warnf(template, args...)
}
func Warnw(msg string, keysAndValues ...interface{}) {
	logger.Sugar().Warnw(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	logger.Sugar().Error(args)
}
func Errorf(template string, args ...interface{}) {
	logger.Sugar().Errorf(template, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Sugar().Errorw(msg, keysAndValues...)
}

func Panic(args ...interface{}) {
	logger.Sugar().Panic(args)
}

func Panicf(template string, args ...interface{}) {
	logger.Sugar().Panicf(template, args...)
}
func Panicw(msg string, keysAndValues ...interface{}) {
	logger.Sugar().Panicw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	logger.Sugar().Fatal(args)
}

func Fatalf(template string, args ...interface{}) {
	logger.Sugar().Fatalf(template, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	logger.Sugar().Fatalw(msg, keysAndValues...)
}
