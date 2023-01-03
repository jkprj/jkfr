package utils

import (
	"reflect"
)

func ResetServerAddr(svrAddr, bindAddr *string) (err error) {

	// 如果只设置了一个，则使服务地址和绑定地址一致
	if "" == *bindAddr {
		*bindAddr = *svrAddr
	}
	if "" == *svrAddr {
		*svrAddr = *bindAddr
	}

	// tmpAddr, svr_port, err := unet.ParseHostAddr(*svrAddr)
	// if "" == tmpAddr {
	// 	jklog.Errorw("unet.ParseHostAddr error", "svrAddr", *svrAddr, "error", "svrAddr IP must not be empty")
	// 	return errors.New("svrAddr IP must not be empty")
	// }

	// var bind_port int = 0
	// *bindAddr, bind_port, err = unet.GetRandomHostAddr(*bindAddr)
	// if nil != err {
	// 	jklog.Errorw("unet.GetRandomHostAddr fail", "bindAddr", *bindAddr, "error", err)
	// 	return err
	// }

	// if 0 == svr_port {
	// 	*svrAddr = tmpAddr + ":" + strconv.Itoa(bind_port)
	// }

	return nil
}

func ZeroStruct(st interface{}) {

	v := reflect.ValueOf(st)

	if reflect.Ptr != v.Kind() || v.IsNil() {
		return
	}

	e := v.Elem()
	if reflect.Struct != e.Kind() {
		return
	}

	for i := 0; i < e.NumField(); i++ {

		fv := e.Field(i)

		if fv.IsZero() || false == fv.CanSet() {
			continue
		}

		fv.Set(reflect.Zero(fv.Type()))
	}
}
