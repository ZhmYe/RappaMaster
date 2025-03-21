package Oracle

import (
	"BHLayer2Node/paradigm"
	"github.com/ethereum/go-ethereum/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PersistedOracle struct {
	channel     *paradigm.RappaChannel
	mySQLConfig paradigm.DBConnection //mysql连接配置
	db          *gorm.DB              //数据库访问对象
}

func (o *PersistedOracle) Start() {
}

func NewPersistedOracle() *PersistedOracle {
	//TODO 初始化数据库连接,这个可以写到config中
	dbConfig := paradigm.DBConnection{
		Username:      "root",
		Password:      "520@111zz",
		Host:          "127.0.0.1",
		Port:          3306,
		Dbname:        "db_rappa",
		Timeout:       "5s",
		IsAutoMigrate: true,
	}

	// 这里使用gorm简化开发，目前打印SQL语句
	dsn := dbConfig.GetDsn()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Error("DB connection is error:", "dsn", dsn, "err", err)
		return nil
	}

	if dbConfig.IsAutoMigrate {
		// 开启自动迁移，将根据模型自动创建和更新数据库表
		err = db.AutoMigrate(&paradigm.Slot{})
		if err != nil {
			log.Error("auto migrate is wrong:", "dsn", dsn, "err", err)
			return nil
		}
	}

	//TODO 从数据库读取record信息

	return &PersistedOracle{
		channel:     nil,
		mySQLConfig: dbConfig,
		db:          db,
	}
}
