package main

import (
	"monitor/utils"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

func cpuMetric() {
	go func() {
		utils.Log().Debug("cpu metric started")

		var err error
		gCpuPhysicalNum, err = cpu.Counts(false)
		if err != nil {
			utils.Log().Debug("get cpu physical number failed:", err)
			return
		}
		gCpuLogicNum, err = cpu.Counts(true)
		if err != nil {
			utils.Log().Debug("get cpu logic number failed:", err)
			return
		}

		for {
			if stop {
				utils.Log().Debug("cpu metric stopped")
				return
			}
			cpuPercentList, err := cpu.Percent(0, false)
			if err != nil {
				utils.Log().Debug("get cpu percent failed:", err)
				return
			}
			gCpuPercent = cpuPercentList[0]

			time.Sleep(time.Second)
		}
	}()
}
