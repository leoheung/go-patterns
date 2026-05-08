package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

func IsDev() bool {
	env := os.Getenv("env")
	isDev := env == "dev"
	return isDev
}

func LogMessage(message string) {
	if IsDev() {
		fmt.Println(message) // 开发环境使用 fmt.Println
	} else {
		log.Println(message) // 生产环境使用 log
	}
}

func DevLogError(errMsg string) {
	PrintlnColor(Red, fmt.Sprintf("[Dev Logs] - %s: %s", time.Now().Format("2006-01-02 15:04:05"), errMsg))
}

func DevLogInfo(infoMsg string) {
	PrintlnColor(BrightBlue, fmt.Sprintf("[Dev Logs] - %s: %s", time.Now().Format("2006-01-02 15:04:05"), infoMsg))
}

func DevLogSuccess(successMsg string) {
	PrintlnColor(Green, fmt.Sprintf("[Dev Logs] - %s: %s", time.Now().Format("2006-01-02 15:04:05"), successMsg))
}
