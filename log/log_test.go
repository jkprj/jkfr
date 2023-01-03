package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	args := ZapOptions{}
	args.MaxAge = 7
	InitLogger(&args)
	//log.InitLogger()
	defer Usync()
	Info("info test......")
	Infof("infof>>>>>>>>%s", "infof infof")
	Debug("debug test.....")
	Debugf("debugf 》》》》》》%s", "debugf degbuf")
	Warn("warning test.......")
	Warnf("warnf>>>>>>>>>%s", "warnf warnf")
	Error("error test.....")
	Errorf("errorf>>>>>>>%s", "errorf errorf")
	log := GetLogger()
	log.Info("555555555555555555")
	log.Info("ttttttttttt5")
}
