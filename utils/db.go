package utils

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"os"
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

func GetDBConnection(host, port, username, dbname, password string) *gorm.DB {
	dbOnce.Do(func() {
		WriteLog("Initialize db connection...", LogLevelInfo)
		connection := "host=" + host + " port=" + port + " user=" + DecryptCred("db-conn", username) + " dbname=" + dbname +
			" password=" + DecryptCred("db-conn", password) + " sslmode=" + os.Getenv("DATABASE_SSL")

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
		db.LogMode(true)

		dbInstance = &dbUtils{
			db: db,
		}
	})

	return dbInstance.db
}
