package config

import "fmt"

type DatabaseConfig struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Dbname       string `json:"dbname"`
	Timeout      string `json:"timeout"`
	MaxIdleConns int    `json:"maxIdleConns"`
	MaxOpenConns int    `json:"maxOpenConns"`
	MaxLifetime  string `json:"maxLifetime"`
}

func (dbConfig *DatabaseConfig) SetDefault() {
	dbConfig.Username = "root"
	dbConfig.Password = "bassword"
	dbConfig.Host = "localhost"
	dbConfig.Port = 3306
	dbConfig.Dbname = "rappa_db"
	dbConfig.Timeout = "5s"
	dbConfig.MaxIdleConns = 10
	dbConfig.MaxOpenConns = 100
	dbConfig.MaxLifetime = "1h"
}

// DSN formats the dsn string with Database config in gorm
func (dbConfig *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%s&multiStatements=true",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Dbname,
		dbConfig.Timeout)
}
