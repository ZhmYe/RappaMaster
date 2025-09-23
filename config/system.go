package config

import "time"

type ComponentConfig struct {
	EpochTimeDuration time.Duration
}

func (cc *ComponentConfig) SetDefault() {
	cc.EpochTimeDuration = 20 * time.Second
}
