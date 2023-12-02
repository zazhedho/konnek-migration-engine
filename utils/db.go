package utils

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type dbUtils struct {
	db *gorm.DB
}

var dbInstance *dbUtils
var dbReportInstance *dbUtils
var dbOnce sync.Once
var dbReportOnce sync.Once

func GetDBConnection() *gorm.DB {
	dbOnce.Do(func() {
		WriteLog("Initialize db connection...", LogLevelInfo)
		connection := "host=" + os.Getenv("DATABASE_HOST") + " port=" + os.Getenv("DATABASE_PORT") + " user=" + DecryptCred("db-conn", os.Getenv("USERNAME_DB")) + " dbname=" + os.Getenv("DATABASE_NAME") +
			" password=" + DecryptCred("db-conn", os.Getenv("PASSWORD_DB")) + " sslmode=" + os.Getenv("DATABASE_SSL")

		//WriteLog(connection, LogLevelInfo)
		db, err := gorm.Open(os.Getenv("DATABASE_TYPE"), connection)
		if err != nil {
			log.Fatalln(connection, err)
			return
		}

		//SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		db.DB().SetMaxIdleConns(10)

		// SetMaxOpenConns sets the maximum number of open connections to the database.
		db.DB().SetMaxOpenConns(150)

		//db.DB().SetConnMaxLifetime(time.Second * 60)
		db.DB().SetConnMaxLifetime(time.Hour)
		db.SingularTable(true)

		logMode, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("DATABASE_LOGMODE")))
		db.LogMode(logMode)

		dbInstance = &dbUtils{
			db: db,
		}
	})

	return dbInstance.db
}

func GetDBNewConnection() *gorm.DB {
	dbOnce.Do(func() {
		WriteLog("Initialize db connection...", LogLevelInfo)
		connection := "host=" + os.Getenv("RE_DATABASE_HOST") + " port=" + os.Getenv("RE_DATABASE_PORT") + " user=" + DecryptCred("db-conn", os.Getenv("RE_USERNAME_DB")) + " dbname=" + os.Getenv("RE_DATABASE_NAME") +
			" password=" + DecryptCred("db-conn", os.Getenv("RE_PASSWORD_DB")) + " sslmode=" + os.Getenv("DATABASE_SSL")

		//WriteLog(connection, LogLevelInfo)
		db, err := gorm.Open(os.Getenv("DATABASE_TYPE"), connection)
		if err != nil {
			log.Fatalln(connection, err)
			return
		}

		//SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		db.DB().SetMaxIdleConns(10)

		// SetMaxOpenConns sets the maximum number of open connections to the database.
		db.DB().SetMaxOpenConns(150)

		//db.DB().SetConnMaxLifetime(time.Second * 60)
		db.DB().SetConnMaxLifetime(time.Hour)
		db.SingularTable(true)

		logMode, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("DATABASE_LOGMODE")))
		db.LogMode(logMode)

		dbInstance = &dbUtils{
			db: db,
		}
	})

	return dbInstance.db
}
