package config

type HTTPConfig struct {
	IP   string // IP
	Port int    // port
}

func (httpConfig *HTTPConfig) SetDefault() {
	httpConfig.IP = "127.0.0.1"
	httpConfig.Port = 8081
}
