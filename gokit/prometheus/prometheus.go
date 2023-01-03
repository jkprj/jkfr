package prometheus

import (
	"sync"
	"time"

	jkregistry "jkfr/gokit/registry"
	"jkfr/gokit/utils"
	jklog "jkfr/log"
	"jkfr/prometheus"
)

var promOnce sync.Once

var regCfgChan chan *jkregistry.RegConfig = make(chan *jkregistry.RegConfig, 100)
var Running bool = false

func init() {

	jkregistry.AppendNotifyConfigInitedHandle(onConfigInitedHandles)

	go func() {
		for regCfg := range regCfgChan {

			if Running {
				continue
			}

			goRunPrometheusSvr(regCfg)
		}
	}()
}

func onConfigInitedHandles(regCfg *jkregistry.RegConfig) {
	if nil == regCfg {
		return
	}
	if "" == regCfg.PrometheusAddr {
		return
	}

	select {
	case regCfgChan <- regCfg:
	default:
		return
	}
}

func goRunPrometheusSvr(regCfg *jkregistry.RegConfig) (err error) {

	errChan := make(chan error, 2)

	go func() {
		err := runPrometheusSvr(regCfg)
		Running = false
		errChan <- err
	}()

	select {
	case err = <-errChan:
	case <-time.After(time.Second):
		Running = true
		err = nil
	}

	if nil != err {
		jklog.Errorw("runPrometheusSvr fail", "PrometheusAddr", regCfg.PrometheusAddr, "err", err.Error())
		Running = false
	} else {
		Running = true
	}

	return err
}

func runPrometheusSvr(regCfg *jkregistry.RegConfig) error {

	err := utils.ResetServerAddr(&regCfg.PrometheusAddr, &regCfg.PrometheusBindAddr)
	if nil != err {
		jklog.Errorw("unet.getPrometheusAddr fail", "PrometheusAddr", regCfg.PrometheusAddr, "PrometheusBindAddr", regCfg.PrometheusBindAddr, "error", err)
		return err
	}

	jklog.Info("PrometheusAddr:", regCfg.PrometheusAddr, ", PrometheusBindAddr:", regCfg.PrometheusBindAddr)

	registry, err := jkregistry.RegistryServer(
		regCfg.PrometheusServerName,
		jkregistry.WithServerAddr(regCfg.PrometheusAddr),
		jkregistry.WithTags(regCfg.ServerName),
		jkregistry.WithConsulAddr(regCfg.ConsulAddr),
		jkregistry.WithBasicAuth(regCfg.UserName, regCfg.Password),
	)
	if nil != err {
		jklog.Errorw("registry prometheusSvr fail", "PrometheusAddr", regCfg.PrometheusAddr, "err", err)
		return err
	}
	defer registry.Deregister()

	err = prometheus.Run(regCfg.PrometheusBindAddr)

	return err
}
