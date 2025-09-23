package config

import "time"

type GrpcConfig struct {
	MessageLimitSize int
	ConnTimeout      time.Duration
	Port             int
}

func (gc *GrpcConfig) SetDefault() {
	gc.ConnTimeout = time.Second * 5
	gc.MessageLimitSize = 1024 * 1024 * 1024
	gc.Port = 8090
}
