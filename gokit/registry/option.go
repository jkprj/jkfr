package registry

import (
	"jkfr/gokit/utils"
	jklog "jkfr/log"
	jkos "jkfr/os"

	consulapi "github.com/hashicorp/consul/api"
)

type RegOption func(cfg *RegConfig) // 服务注册/发现选项
type CallBackFunc func(regCfg *RegConfig, err error)

var callbackFunc CallBackFunc = nil

// consul健康检查服务启动失败默认回调函数
func defaultHealthCheckServerErrorCallback(regCfg *RegConfig, err error) {
	jklog.Errorw("HealthCheckServerError", "regCfg", regCfg, "err", err)
}

type registry struct {
	RegCfg *RegConfig `json:"Registry" toml:"Registry"`
}

type RegConfig struct {
	ServerName string // 服务名称
	SvcHost    string // 服务绑定ip
	SvcPort    int    // 服务绑定端口

	ServerAddr string `json:"ServerAddr" toml:"ServerAddr"` // 服务地址
	ConsulAddr string `json:"ConsulAddr" toml:"ConsulAddr"` // consul服务地址
	Namespace  string `json:"Namespace" toml:"Namespace"`   // 注册consul时的Namespace，非企业版，似乎无效
	UserName   string `json:"UserName" toml:"UserName"`     // consul服务的用户名(如果需要)
	Password   string `json:"Password" toml:"Password"`     // consul服务的密码(如果需要)

	HealthCheckAddr                string `json:"HealthCheckAddr" toml:"HealthCheckAddr"`                               // consul健康检查服务地址
	HealthCheckBindAddr            string `json:"HealthCheckBindAddr" toml:"HealthCheckBindAddr"`                       // consul健康检查绑定服务地址
	DeregisterCriticalServiceAfter int    `json:"DeregisterCriticalServiceAfter" toml:"DeregisterCriticalServiceAfter"` // consul健康检查Criticald多久后取消注册
	HealthCheckInterval            int    `json:"HealthCheckInterval" toml:"HealthCheckInterval"`                       // consul健康检查间隔时间
	HealthCheckTimeOut             int    `json:"HealthCheckTimeOut" toml:"HealthCheckTimeOut"`                         // consul健康检查超时时间

	ConsulTags []string `json:"ConsulTags" toml:"ConsulTags"` // 注册到consul的tags

	// 注册等待超时时间
	WaitTime int `json:"WaitTime" toml:"WaitTime"` // 注册时，等待响应超时时间

	PrometheusNameSpace  string `json:"PrometheusNameSpace" toml:"PrometheusNameSpace"`   // prometheus 添加监控时的名称
	PrometheusServerName string `json:"PrometheusServerName" toml:"PrometheusServerName"` // prometheus 服务名称

	// 服务地址，绑定地址主要考虑当绑定所有地址时，服务地址作为指定注册时的地址
	PrometheusAddr     string `json:"PrometheusAddr" toml:"PrometheusAddr"`         // prometheus 服务地址
	PrometheusBindAddr string `json:"PrometheusBindAddr" toml:"PrometheusBindAddr"` // prometheus 绑定的地址

	ConfigPath  string                  // 配置文件路径
	PassingOnly bool                    `json:"PassingOnly" toml:"PassingOnly"` // 服务发现时是否只获取正常的服务信息
	QueryOpts   *consulapi.QueryOptions // 服务发现选项
}

// 读取配置
func defaultconfig(name string) *RegConfig {
	cfg := new(RegConfig)

	// 从环境变量读取配置，未配置环境变量则设置默认值
	cfg.ServerName = name
	cfg.ServerAddr = jkos.GetEnvString("R_SERVER_ADDR", "")
	cfg.ConsulAddr = jkos.GetEnvString("R_CONSUL_ADDR", "127.0.0.1:8500")
	cfg.HealthCheckAddr = jkos.GetEnvString("R_HEALTH_CHECK_ADDR", "127.0.0.1")
	cfg.HealthCheckInterval = jkos.GetEnvInt("R_HEALTH_CHECK_INTERVAL", 1)
	cfg.HealthCheckTimeOut = jkos.GetEnvInt("R_HEALTH_CHECK_TIMEOUT", 60)
	cfg.DeregisterCriticalServiceAfter = jkos.GetEnvInt("R_DEREGISTER_CRITICAL_SERVICE_AFTER", 30)
	cfg.PassingOnly = jkos.GetEnvBool("R_PASSING_ONLY", true)
	cfg.Namespace = jkos.GetEnvString("R_NAMESPACE", "")
	cfg.PrometheusNameSpace = jkos.GetEnvString("R_PROMETHEUS_NAMESPACE", name)
	cfg.PrometheusAddr = jkos.GetEnvString("R_PROMETHEUS_ADDR", "")
	cfg.PrometheusBindAddr = jkos.GetEnvString("R_PROMETHEUS_BIND_ADDR", "")
	cfg.PrometheusServerName = jkos.GetEnvString("R_PROMETHEUS_SERVER_NAME", "prometheus")
	cfg.UserName = jkos.GetEnvString("R_USERNAME", "")
	cfg.Password = jkos.GetEnvString("R_PASSWORD", "")
	cfg.ConsulTags = jkos.GetEnvStrings("R_CONSUL_TAGS", ",", nil)
	cfg.WaitTime = jkos.GetEnvInt("R_WAIT_TIME", 60)

	// 从环境变量读取配置文件路径，然后读取配置文件
	cfg.ConfigPath = jkos.GetEnvString("R_CONFIG_PATH", "")
	if jkos.IsFileExists(cfg.ConfigPath) {
		WithFile(cfg.ConfigPath)(cfg)
	}

	loadDefaultServerConfig(name, cfg)

	if nil == callbackFunc {
		callbackFunc = defaultHealthCheckServerErrorCallback
	}

	return cfg
}

// 尝试读取默认路径的配置文件配置
func loadDefaultServerConfig(name string, cfg *RegConfig) {

	fileName := jkos.CurDir() + "/conf/" + name

	conf := fileName + ".toml"
	if jkos.IsFileExists(conf) {
		WithFile(conf)(cfg)
	}

	conf = fileName + ".json"
	if jkos.IsFileExists(conf) {
		WithFile(conf)(cfg)
	}

	conf = jkos.CurDir() + "/conf/registry.json"
	if jkos.IsFileExists(conf) {
		WithFile(conf)(cfg)
	}

	conf = jkos.CurDir() + "/conf/registry.toml"
	if jkos.IsFileExists(conf) {
		WithFile(conf)(cfg)
	}
}

// 设置consul健康检查服务启动失败默认回调函数
func SetRegistryHealthCheckServerCallBackFunc(cbFunc CallBackFunc) {
	callbackFunc = cbFunc
}

// 服务地址
func WithServerAddr(addr string) RegOption {
	return func(cfg *RegConfig) {
		cfg.ServerAddr = addr
	}
}

// consul服务地址
func WithConsulAddr(addr string) RegOption {
	return func(cfg *RegConfig) {
		cfg.ConsulAddr = addr
	}
}

// 注册consul时的Namespace，非企业版，似乎无效
func WithNamespace(namespace string) RegOption {
	return func(cfg *RegConfig) {
		cfg.Namespace = namespace
	}
}

// consul服务的用户名,密码(如果需要)
func WithBasicAuth(userName, password string) RegOption {
	return func(cfg *RegConfig) {
		cfg.UserName = userName
		cfg.Password = password
	}
}

// consul健康检查绑定服务地址
func WithHealthCheckAddr(addr string) RegOption {
	return func(cfg *RegConfig) {
		cfg.HealthCheckAddr = addr
	}
}

// consul健康检查--Critical后多久删除注册
func WithDeregisterCriticalServiceAfter(deregisterCriticalServiceAfter int) RegOption {
	return func(cfg *RegConfig) {
		cfg.DeregisterCriticalServiceAfter = deregisterCriticalServiceAfter
	}
}

// consul健康检查间隔时间
func WithHealthCheckInterval(interval int) RegOption {
	return func(cfg *RegConfig) {
		cfg.HealthCheckInterval = interval
	}
}

// consul健康检查超时时间
func WithHealthCheckTimeOut(timeout int) RegOption {
	return func(cfg *RegConfig) {
		cfg.HealthCheckTimeOut = timeout
	}
}

// 注册到consul的tags
func WithTags(tags ...string) RegOption {
	return func(cfg *RegConfig) {
		cfg.ConsulTags = tags
	}
}

// 注册时，等待响应超时时间
func WithWaitTime(waitTime int) RegOption {
	return func(cfg *RegConfig) {
		cfg.WaitTime = waitTime
	}
}

// prometheus 添加监控时的名称
func WithPrometheusNameSpace(prometheusNameSpace string) RegOption {
	return func(cfg *RegConfig) {
		cfg.PrometheusNameSpace = prometheusNameSpace
	}
}

// prometheus 服务地址
func WithPrometheusAddr(prometheusAddr string) RegOption {
	return func(cfg *RegConfig) {
		cfg.PrometheusAddr = prometheusAddr
	}
}

// prometheus 绑定地址
func WithPrometheusBindAddr(prometheusBindAddr string) RegOption {
	return func(cfg *RegConfig) {
		cfg.PrometheusBindAddr = prometheusBindAddr
	}
}

// prometheus 服务名称
func WithPrometheusServerName(prometheusServerName string) RegOption {
	return func(cfg *RegConfig) {
		cfg.PrometheusServerName = prometheusServerName
	}
}

// 服务发现时是否只获取正常的服务信息
func WithPassingOnly(passingOnly bool) RegOption {
	return func(cfg *RegConfig) {
		cfg.PassingOnly = passingOnly
	}
}

// 服务发现选项
func WithQueryOptions(queryOpts *consulapi.QueryOptions) RegOption {
	return func(cfg *RegConfig) {
		cfg.QueryOpts = queryOpts
	}
}

// 配置文件路径
func WithFile(cfgPath string) RegOption {
	return func(cfg *RegConfig) {
		cfg.ConfigPath = cfgPath

		reg := registry{RegCfg: cfg}
		utils.ReadConfigFile(cfgPath, &reg)
	}
}
