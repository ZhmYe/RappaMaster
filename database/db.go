package database

import (
	"RappaMaster/config"
	"RappaMaster/paradigm"
	"fmt"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"path"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DatabaseService manages db(mysql now)
type DatabaseService struct {
	config.DatabaseConfig
	db *gorm.DB
}

func (dbs *DatabaseService) IsCreated() bool {
	return dbs.db != nil
}
func (dbs *DatabaseService) script(path string, values ...interface{}) error {
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	sqlStr := string(sqlBytes)
	if err = dbs.db.Exec(sqlStr, values...).Error; err != nil {
		return paradigm.RaiseError(paradigm.DatabaseError, fmt.Sprintf("Database executes scripts failed"), err)
	}
	return nil
}
func (dbs *DatabaseService) Init() error {
	if dbs.IsCreated() {
		return nil
	}
	fmt.Println(dbs.DSN())
	logFile, err := os.OpenFile(path.Join(config.ProjectRootPath, "logs/slow_sql.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return paradigm.RaiseError(paradigm.FileError, "Create slow sql log failed", err)
	}
	defer logFile.Close()

	customLogger := logger.New(
		log.New(logFile, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
	gormDB, err := gorm.Open(mysql.Open(dbs.DSN()), &gorm.Config{Logger: customLogger})
	if err != nil {
		return paradigm.RaiseError(paradigm.DatabaseError, "Database inits failed", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return paradigm.RaiseError(paradigm.DatabaseError, "Get database failed", err)
	}
	sqlDB.SetMaxIdleConns(dbs.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbs.MaxOpenConns)
	maxLifetime, err := time.ParseDuration(dbs.MaxLifetime)
	if err != nil {
		return paradigm.RaiseError(paradigm.ValueError, fmt.Sprintf("Parse MaxLifeTime %s failed", dbs.MaxLifetime), err)
	}
	sqlDB.SetConnMaxLifetime(maxLifetime)
	dbs.db = gormDB
	return dbs.script("./sql/schema.sql")

}

func NewDatabaseService(config config.DatabaseConfig) *DatabaseService {
	return &DatabaseService{
		DatabaseConfig: config,
		db:             nil,
	}
}

//
//func (dbs *DatabaseService) AutoMigrate() error {
//	return dbs.db.AutoMigrate(
//		&paradigm.Slot{},
//		&paradigm.Task{},
//		&paradigm.DevEpoch{},
//		&paradigm.DevReference{},
//		&Date.DateRecord{},
//	)
//}
//
//func (o DatabaseService) TruncateAll() error {
//	tables := []string{"date_records", "dev_epoches", "dev_references", "slots", "tasks"}
//	for _, table := range tables {
//		if err := o.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table)).Error; err != nil {
//			return fmt.Errorf("failed to truncate table %s: %v", table, err)
//		}
//	}
//	return nil
//}
