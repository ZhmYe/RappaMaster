package Database

import (
	"BHLayer2Node/Date"
	"BHLayer2Node/paradigm"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseService struct {
	db     *gorm.DB
	config *paradigm.BHLayer2NodeConfig
}

// 全局单例
//var (
//	dbService     *DatabaseService
//	dbServiceOnce sync.Once
//)

func NewDatabaseService(config *paradigm.BHLayer2NodeConfig) (*DatabaseService, error) {
	var initErr error
	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%s",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Dbname,
		config.Database.Timeout)

	// 初始化GORM
	gormConfig := gorm.Config{}
	if config.DEBUG {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}
	gormDB, err := gorm.Open(mysql.Open(dsn), &gormConfig)
	if err != nil {
		initErr = fmt.Errorf("failed to open database: %v", err)
		return nil, err
	}

	// 配置连接池
	sqlDB, err := gormDB.DB()
	if err != nil {
		initErr = fmt.Errorf("failed to get sql.DB: %v", err)
		return nil, err
	}
	sqlDB.SetMaxIdleConns(config.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Database.MaxOpenConns)
	maxLifetime, _ := time.ParseDuration(config.Database.MaxLifetime)
	sqlDB.SetConnMaxLifetime(maxLifetime)

	// 自动迁移
	if config.Database.IsAutoMigrate {
		if err := autoMigrate(gormDB); err != nil {
			initErr = fmt.Errorf("auto migrate failed: %v", err)
			return nil, err
		}
	}

	return &DatabaseService{
		db:     gormDB,
		config: config,
	}, initErr
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&paradigm.Slot{},
		&paradigm.Task{},
		&paradigm.DevEpoch{},
		&paradigm.DevReference{},
		&Date.DateRecord{},
	)
}
