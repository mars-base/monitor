package main

import (
	"fmt"
	"monitor/utils"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// global state
var (
	stop    bool = false
	gTicker *utils.RepeatingTimer

	enableCpu     bool
	enableMem     bool
	enableNet     bool
	enableDisk    bool
	enableAll     bool
	interval      int
	interfaceName string

	gCpuPercent     float64
	gCpuLogicNum    int
	gCpuPhysicalNum int
	gMemTotal       uint64
	gMemUsed        uint64
	gMemFree        uint64
	gMemPercent     float64
)

var (
	netMu    sync.RWMutex
	gNetInfo = make(map[string]NetInfo)
	gNetNames []string
)

var (
	diskMu    sync.RWMutex
	gDiskInfo = make(map[string]DiskInfo)
	gDiskNames []string
)

// NetInfo holds per-interface network speed data
type NetInfo struct {
	Sent        float64 // send rate (b/s)
	Recv        float64 // recv rate (b/s)
	PacketsSent float64 // send packet rate (pps)
	PacketsRecv float64 // recv packet rate (pps)
}

// DiskInfo holds per-device disk throughput data
type DiskInfo struct {
	Read  uint64 // read bytes/s
	Write uint64 // write bytes/s
}

// cobra commands
var rootCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Simple tool to monitor cpu/mem/network.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of monitor",
	Run: func(cmd *cobra.Command, args []string) {
		utils.Log().Debug(AppName, AppVersion)
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show command line parameters",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("enable cpu:", enableCpu)
		fmt.Println("enable mem:", enableMem)
		fmt.Println("enable net:", enableNet)
		fmt.Println("enable disk:", enableDisk)
		fmt.Println("enable all:", enableAll)
		fmt.Printf("interval: %d seconds\n", interval)
		fmt.Println("interface name:", interfaceName)
	},
}

// CmdEntryPoint is the entry point of cobra command, called in main.go
func CmdEntryPoint() {
	rootCmd.PersistentFlags().BoolVarP(&enableCpu, "cpu", "c", false, "enable cpu metric")
	rootCmd.PersistentFlags().BoolVarP(&enableMem, "mem", "m", false, "enable mem metric")
	rootCmd.PersistentFlags().BoolVarP(&enableNet, "net", "n", false, "enable net metric")
	rootCmd.PersistentFlags().BoolVarP(&enableDisk, "disk", "d", false, "enable disk metric")
	rootCmd.PersistentFlags().BoolVarP(&enableAll, "all", "a", false, "enable all metrics")
	rootCmd.PersistentFlags().IntVarP(&interval, "interval", "i", 2, "interval in seconds")
	rootCmd.PersistentFlags().StringVarP(&interfaceName, "if", "", "", "network interface name")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run monitor tool",
	Run: func(cmd *cobra.Command, args []string) {
		utils.Log().Debug("enable cpu:", enableCpu)
		utils.Log().Debug("enable mem:", enableMem)
		utils.Log().Debug("enable net:", enableNet)
		utils.Log().Debug("enable disk:", enableDisk)
		utils.Log().Debug("enable all:", enableAll)
		utils.Log().Debug("interval seconds:", interval)
		utils.Log().Debug("interface name:", interfaceName)

		if enableAll {
			enableCpu = true
			enableMem = true
			enableNet = true
			enableDisk = true
		}
		if !enableCpu && !enableMem && !enableNet && !enableDisk {
			logger.Println("no monitor module enabled, -h for help")
			return
		}

		utils.Log().Debug("start monitor module")
		if enableCpu {
			cpuMetric()
		}
		if enableMem {
			memMetric()
		}
		if enableNet {
			netMetric()
		}
		if enableDisk {
			diskMetric()
		}

		// start display loop
		handleMetric()

		// wait for stop signal
		utils.AddGracefulExit(func() error {
			stop = true
			stopHandleMetric()
			time.Sleep(time.Second)
			return nil
		})
	},
}
