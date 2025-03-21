package paradigm

import "fmt"

type DBConnection struct {
	Username      string // 账号
	Password      string // 密码
	Host          string // 数据库地址，可以是IP或者域名
	Port          int    // 数据库端口
	Dbname        string // 数据库名
	Timeout       string // 连接超时时间
	IsAutoMigrate bool   // 是否开启自动迁移
}

func (d *DBConnection) GetDsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local&timeout=%s",
		d.Username, d.Password, d.Host, d.Port, d.Dbname, d.Timeout)
}
