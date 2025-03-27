package database

import (
	"BHLayer2Node/Date"
	"BHLayer2Node/paradigm"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseService struct {
	db      *gorm.DB
	channel *paradigm.RappaChannel
	mu      sync.RWMutex
	config  *paradigm.BHLayer2NodeConfig
}

// 全局单例
var (
	dbService     *DatabaseService
	dbServiceOnce sync.Once
)

func NewDatabaseService(channel *paradigm.RappaChannel, config *paradigm.BHLayer2NodeConfig) (*DatabaseService, error) {
	var initErr error
	dbServiceOnce.Do(func() {
		// 构建DSN
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%s",
			config.Database.Username,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Dbname,
			config.Database.Timeout)

		// 初始化GORM
		gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err != nil {
			initErr = fmt.Errorf("failed to open database: %v", err)
			return
		}

		// 配置连接池
		sqlDB, err := gormDB.DB()
		if err != nil {
			initErr = fmt.Errorf("failed to get sql.DB: %v", err)
			return
		}
		sqlDB.SetMaxIdleConns(config.Database.MaxIdleConns)
		sqlDB.SetMaxOpenConns(config.Database.MaxOpenConns)
		maxLifetime, _ := time.ParseDuration(config.Database.MaxLifetime)
		sqlDB.SetConnMaxLifetime(maxLifetime)

		// 自动迁移
		if config.Database.IsAutoMigrate {
			if err := autoMigrate(gormDB); err != nil {
				initErr = fmt.Errorf("auto migrate failed: %v", err)
				return
			}
		}

		// 启动健康检查协程
		go startHealthCheck(sqlDB)

		dbService = &DatabaseService{
			db:      gormDB,
			channel: channel,
			config:  config,
		}
	})
	return dbService, initErr
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

func startHealthCheck(sqlDB *sql.DB) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := sqlDB.Ping(); err != nil {
			log.Printf("⚠️ Database health check failed: %v", err)
		}
	}
}

// 获取数据库实例（线程安全）
func GetDB() *gorm.DB {
	if dbService == nil {
		panic("database service not initialized")
	}
	dbService.mu.RLock()
	defer dbService.mu.RUnlock()
	return dbService.db
}

// 获取原始连接池（用于监控）
func GetSQLDB() *sql.DB {
	sqlDB, err := GetDB().DB()
	if err != nil {
		panic(err)
	}
	return sqlDB
}

// 从数据库里获取最大的EpochID
func GetMaxEpochID() (int32, error) {
	var maxEpochID int32
	result := GetDB().Model(&paradigm.DevEpoch{}).Select("COALESCE(MAX(epoch_id), -1)").Scan(&maxEpochID)
	if result.Error != nil {
		return -1, result.Error
	}
	return maxEpochID, nil
}
