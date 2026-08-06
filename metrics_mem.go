package main

import (
	"monitor/utils"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

func memMetric() {
	go func() {
		utils.Log().Debug("mem metric started")
		for {
			if stop {
				utils.Log().Debug("mem metric stopped")
				return
			}
			memInfo, err := mem.VirtualMemory()
			if err != nil {
				utils.Log().Debug("get mem info failed:", err)
				return
			}
			gMemTotal = memInfo.Total
			gMemUsed = memInfo.Used
			gMemFree = memInfo.Free
			gMemPercent = memInfo.UsedPercent

			time.Sleep(time.Second)
		}
	}()
}
