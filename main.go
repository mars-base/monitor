package main

// Time    :   2025-11-25 05:19:26 PM
// Author  :   diwen

import (
	"monitor/utils" // private package
)

var (
	logger = utils.Logger
)

func main() {
	utils.SetLogLevel("Info")
	// utils.SetLogLevel("Debug")
	// utils.SetLogUseJson(false)
	utils.SetLogToFile(true, "x.log", 1)
	utils.InitLog()
	utils.Log(utils.LogFields{}).Debug("Monitor tool. -h for help")
	logger.Println("Monitor tool. -h for help")

	CmdEntryPoint()
}
