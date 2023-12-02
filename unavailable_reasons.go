package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/joho/godotenv/autoload"
	"konnek-migration/utils"
	"os"
)

func main() {
	// Create source DB connection
	scDB := utils.GetDBConnection(os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("USERNAME_DB"), os.Getenv("DATABASE_NAME"), os.Getenv("PASSWORD_DB"))
	defer func(scDB *gorm.DB) {
		err := scDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(scDB)

	// Create destination DB connection
	dstDB := utils.GetDBConnection(os.Getenv("RE_DATABASE_HOST"), os.Getenv("RE_DATABASE_PORT"), os.Getenv("RE_USERNAME_DB"), os.Getenv("RE_DATABASE_NAME"), os.Getenv("RE_PASSWORD_DB"))
	defer func(dstDB *gorm.DB) {
		err := dstDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection scDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(dstDB)
}
