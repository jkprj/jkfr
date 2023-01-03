package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:    "Msg",
		LevelKey:      "Level",
		TimeKey:       "Time",
		NameKey:       "logger",
		CallerKey:     "File",
		StacktraceKey: "trace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     nil,
	}
}
func initZapLog(ops *ZapOptions) {
	out_info := zapcore.AddSync(&lumberjack.Logger{
		Filename:   ops.LogPath + "/" + ops.FileName + ".log",
		MaxSize:    ops.MaxSize,   //大小M
		MaxAge:     ops.MaxAge,    //天
		MaxBackups: ops.MaxBackup, //最多保留3个备份
		LocalTime:  true,
		Compress:   false,
	})
	out_error := zapcore.AddSync(&lumberjack.Logger{
		Filename:   ops.LogPath + "/" + ops.FileName + ".err",
		MaxSize:    ops.MaxSize,   //大小
		MaxAge:     ops.MaxAge,    //天
		MaxBackups: ops.MaxBackup, //最多保留3个备份
		LocalTime:  true,
		Compress:   false,
	})
	var loglevel zapcore.Level
	if level, ok := LogLevel[ops.LogLevel]; ok {
		loglevel = level
	} else {
		loglevel = zap.InfoLevel
	}
	core = zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(NewEncoderConfig()), zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout), out_info), loglevel),
		zapcore.NewCore(zapcore.NewJSONEncoder(NewEncoderConfig()), zapcore.NewMultiWriteSyncer(
			out_error), zap.WarnLevel),
	)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

func Usync() {
	logger.Sync()
}

func GetLogger() *zap.Logger {
	return logger.WithOptions(zap.AddCallerSkip(-1))
}
