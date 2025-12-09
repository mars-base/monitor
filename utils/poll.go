package utils

import (
	"github.com/wangdiwen/gopool"
)

// default queue size 1000000
// pollNum is maxWorkers == minWorkers
func InitPool(pollNum int) gopool.GoPool {
	return gopool.NewGoPool(pollNum)
}
