package log

import (
	kitlog "github.com/go-kit/kit/log"
	kitzap "github.com/go-kit/kit/log/zap"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

/***************************************** global value ***************************************************/

var DebugwLogger kitlog.Logger
var InfowLogger kitlog.Logger
var WarnwLogger kitlog.Logger
var ErrorwLogger kitlog.Logger
var DPanicwLogger kitlog.Logger
var PanicwLogger kitlog.Logger
var FatalwLogger kitlog.Logger

/***************************************** struct ********************************************************/

func InitLogger(logHandle *zap.Logger) {

	if nil == logHandle {
		return
	}

	logger = logHandle.WithOptions(zap.AddCallerSkip(-1))

	DebugwLogger = kitzap.NewZapSugarLogger(logger, zapcore.DebugLevel)
	InfowLogger = kitzap.NewZapSugarLogger(logger, zapcore.InfoLevel)
	WarnwLogger = kitzap.NewZapSugarLogger(logger, zapcore.WarnLevel)
	ErrorwLogger = kitzap.NewZapSugarLogger(logger, zapcore.ErrorLevel)
	DPanicwLogger = kitzap.NewZapSugarLogger(logger, zapcore.DPanicLevel)
	PanicwLogger = kitzap.NewZapSugarLogger(logger, zapcore.PanicLevel)
	FatalwLogger = kitzap.NewZapSugarLogger(logger, zapcore.FatalLevel)
}
