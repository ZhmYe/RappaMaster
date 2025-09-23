package config

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

func (rc *RedisConfig) SetDefault() {
	rc.Addr = "127.0.0.1:6379"
	rc.Password = ""
	rc.DB = 0
}
