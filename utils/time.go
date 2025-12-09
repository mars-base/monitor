package utils

import (
	"time"
)

func GetNowTime() string {
	return time.Now().Format("2006-01-02 03:04:05")
}

func SleepSeconds(interval int) {
	time.Sleep(time.Duration(interval) * time.Second)
}

func SleepMilliseconds(interval int) {
	time.Sleep(time.Duration(interval) * time.Millisecond)
}
