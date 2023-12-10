package utils

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func Init() {
	envFilePath1 := "../../.env"

	// Load file .env pertama dari path yang berbeda
	err := godotenv.Load(envFilePath1)
	if err != nil {
		log.Fatal("Error loading first .env file:", err)
	}

	// Ganti dengan path yang sesuai dengan lokasi file .env kedua Anda
	envFilePath2 := ".env"

	// Load file .env kedua dari path yang berbeda
	err = godotenv.Load(envFilePath2)
	if err != nil {
		log.Fatal("Error loading second .env file:", err)
	}
}

func GetMyIP() string {
	myAddr := "unknown"
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				myAddr = ipnet.IP.String()
				break
			}
		}
	}

	myAddr += strings.Repeat(" ", 15-len(myAddr))

	return myAddr
}

const (
	LogLevelPanic = 0
	LogLevelError = 1
	LogLevelFail  = 2
	LogLevelInfo  = 3
	LogLevelData  = 4
	LogLevelDebug = 5
)

func WriteLog(msg string, level int) {
	if logLevel, _ := strconv.Atoi(os.Getenv("LOG_LEVEL")); logLevel < level {
		return
	}

	logTime := time.Now().Format("2006/01/02 15:04:05.000000")
	logMcs := time.Now().Format(".000000")

	fileName := "konnek_migration_" + time.Now().Format("2006_01_02")
	if os.Getenv("APP_NAME") != "" {
		fileName = os.Getenv("APP_NAME") + "_" + time.Now().Format("2006_01_02")
	}

	// Membuat file log
	logFile, err := os.OpenFile("../../log/"+fileName+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	switch level {
	case LogLevelPanic:
		log.Printf("%s [%s][PANIC]%s\n\n", logMcs, GetMyIP(), msg)
	case LogLevelError:
		log.Printf("%s [%s][ERROR]%s\n\n", logMcs, GetMyIP(), msg)
	case LogLevelFail:
		log.Printf("%s [%s][FAIL ]%s\n\n", logTime, GetMyIP(), msg)
	case LogLevelInfo:
		log.Printf("%s [%s][INFO ]%s\n\n", logTime, GetMyIP(), msg)
	case LogLevelData:
		log.Printf("%s [%s][DATA ]%s\n\n", logTime, GetMyIP(), msg)
	case LogLevelDebug:
		log.Printf("%s [%s][DEBUG]%s\n\n", logTime, GetMyIP(), msg)
	}

}

func WriteErrorMap(filename string, msg interface{}) {
	filename += ".json"
	jsonFile, err := os.OpenFile("../../data/"+filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()
	//_, err = jsonFile.WriteString(msg + "\n")
	//if err != nil {
	//	WriteLog(fmt.Sprintf("failed write to %s; error: %v", filename, err), LogLevelError)
	//}
	encoder := json.NewEncoder(jsonFile)
	err = encoder.Encode(msg)
	if err != nil {
		WriteLog(fmt.Sprintf("failed write to data/%s; error: %v", filename, err), LogLevelError)
	}
}

func WriteToFile(filename string, msg interface{}) {
	var file *os.File
	var err error

	file, err = os.OpenFile("../../data/"+filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		dir := "../data/"
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		file, err = os.OpenFile(dir+filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("%s\n", msg))
	if err != nil {
		WriteLog(fmt.Sprintf("failed write to data/%s; error: %v", filename, err), LogLevelError)
	}
}

const (
	LayoutDate        = "2006-01-02"
	LayoutTime        = "15:04:05"
	LayoutDateTime    = "2006-01-02 15:04:05"
	LayoutDateTimeDot = "2006-01-02 15.04.05"
	LayoutTimestamp   = "2006-01-02 15:04:05.9999999999"
	LayoutDateTimeH   = "2006-01-02 15"
)
