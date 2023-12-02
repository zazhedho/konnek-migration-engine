package utils

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

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

	switch level {
	case LogLevelPanic:
		log.Printf("%s [%s][PANIC]%s\n\n", logMcs, GetMyIP(), msg)
	case LogLevelError:
		log.Printf("%s [%s][ERROR]%s\n\n", logMcs, GetMyIP(), msg)
	case LogLevelFail:
		fmt.Printf("%s [%s][FAIL ]%s\n\n", logTime, GetMyIP(), msg)
	case LogLevelInfo:
		fmt.Printf("%s [%s][INFO ]%s\n\n", logTime, GetMyIP(), msg)
	case LogLevelData:
		fmt.Printf("%s [%s][DATA ]%s\n\n", logTime, GetMyIP(), msg)
	case LogLevelDebug:
		fmt.Printf("%s [%s][DEBUG]%s\n\n", logTime, GetMyIP(), msg)
	}
}
