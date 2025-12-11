package main

import (
	"fmt"
	"monitor/utils"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/spf13/cobra"
)

// global variable
var (
	stop    bool = false
	gTicker *utils.RepeatingTimer

	enableCpu bool
	enableMem bool
	enableNet bool
	interval  int

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
	gNetInfo = make(map[string]struct {
		Sent        float64 // 发送速率（b/s）
		Recv        float64 // 接收速率（b/s）
		PacketsSent float64 // 发送包速率（pps）
		PacketsRecv float64 // 接收包速率（pps）
	})
	gNetNames = []string{}
)

// cobra root command
var rootCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Simple tool to monitor cpu/mem/network.",
	// Long:  `A Fast and Flexible Static Site Generator built with love by spf13 and friends in Go.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of monitor",
	// Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.Log().Debug(gName, gVersion)
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show command line parameters",
	Run: func(cmd *cobra.Command, args []string) {
		// enable flag
		utils.Log().Debug("enable cpu:", enableCpu)
		utils.Log().Debug("enable mem:", enableMem)
		utils.Log().Debug("enable net:", enableNet)
		utils.Log().Debugf("interval: %d seconds", interval)
	},
}

// CmdEntryPoint is the entry point of cobra command,
// it should be called in main.go
func CmdEntryPoint() {
	// add flag
	rootCmd.PersistentFlags().BoolVarP(&enableCpu, "cpu", "c", false, "enable cpu monitor")
	rootCmd.PersistentFlags().BoolVarP(&enableMem, "mem", "m", false, "enable mem monitor")
	rootCmd.PersistentFlags().BoolVarP(&enableNet, "net", "n", false, "enable net monitor")
	rootCmd.PersistentFlags().IntVarP(&interval, "interval", "i", 2, "monitor interval in seconds")

	// add command
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(runCmd)

	// 执行命令行
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
		// show parameters
		utils.Log().Debug("enable cpu:", enableCpu)
		utils.Log().Debug("enable mem:", enableMem)
		utils.Log().Debug("enable net:", enableNet)
		utils.Log().Debug("interval seconds:", interval)

		if enableCpu || enableMem || enableNet {
			utils.Log().Debug("start monitor module")
			if enableCpu {
				utils.Log().Debug("start cpu monitor module")
				cpuMetric()
			}
			if enableMem {
				utils.Log().Debug("start mem monitor module")
				memMetric()
			}
			if enableNet {
				utils.Log().Debug("start net monitor module")
				netMetric()
			}
		} else {
			logger.Println("no monitor module enabled, -h for help")
			return
		}

		// start handle metric
		handleMetric()

		// hang main thread and wait for stop signal
		utils.AddGracefulExit(func() error {
			// stop handle metric
			stop = true
			// stop ticker
			stopHandleMetric()
			time.Sleep(time.Second)
			return nil
		}) // finish auto exit
	},
}

func cpuMetric() {
	go func() {
		utils.Log().Debug("cpu metric started")

		// get cpu number
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
			utils.Log().Debug("collect cpu percent")
			cpuPercentList, err := cpu.Percent(0, false)
			if err != nil {
				utils.Log().Debug("get cpu percent failed:", err)
				return
			}
			// store cpu percent to global variable
			gCpuPercent = cpuPercentList[0]

			time.Sleep(time.Second)
		}
	}()
}

func memMetric() {
	go func() {
		utils.Log().Debug("mem metric started")
		for {
			if stop {
				utils.Log().Debug("mem metric stopped")
				return
			}
			utils.Log().Debug("collect mem info")
			memInfo, err := mem.VirtualMemory()
			if err != nil {
				utils.Log().Debug("get mem info failed:", err)
				return
			}
			// store mem info to global variable
			gMemTotal = memInfo.Total
			gMemUsed = memInfo.Used
			gMemFree = memInfo.Free
			gMemPercent = memInfo.UsedPercent

			time.Sleep(time.Second)
		}
	}()
}

func netMetric() {
	// TODO: implement net metric
	go func() {
		utils.Log().Debug("net metric started")

		var netHistory = make(map[string]struct {
			prevSent, prevRecv, prevPacketsSent, prevPacketsRecv uint64
			firstRun                                             bool
		})

		// get net interface name
		netIfs, err := net.Interfaces()
		if err != nil {
			utils.Log().Errorf("failed to get net interfaces: %v", err)
			return
		}

		for _, netIf := range netIfs {
			utils.Log().Debugf("net interface: %s", netIf.Name)
			// 初始化网络接口信息
			gNetInfo[netIf.Name] = struct {
				Sent        float64
				Recv        float64
				PacketsSent float64
				PacketsRecv float64
			}{
				Sent:        0,
				Recv:        0,
				PacketsSent: 0,
				PacketsRecv: 0,
			}
			// 记录网络接口名称
			gNetNames = append(gNetNames, netIf.Name)

			// 初始化历史记录
			netHistory[netIf.Name] = struct {
				prevSent, prevRecv, prevPacketsSent, prevPacketsRecv uint64
				firstRun                                             bool
			}{
				prevSent:        0,
				prevRecv:        0,
				prevPacketsSent: 0,
				prevPacketsRecv: 0,
				firstRun:        true,
			}
		}

		utils.Log().Debugf("net interface count: %d", len(netIfs))

		for {
			if stop {
				utils.Log().Debug("net metric stopped")
				return
			}
			utils.Log().Debug("collect net info")
			netIO, err := net.IOCounters(true) // true表示分别获取每个网卡的数据
			if err != nil {
				utils.Log().Errorf("failed to get net io counters: %v", err)
				return
			}
			// 遍历每个网络接口
			for _, counter := range netIO {
				// 记录网络统计信息
				utils.Log().Debugf("Net interface: %s", counter.Name)
				utils.Log().Debugf("  Sent: %d Bytes", counter.BytesSent)
				utils.Log().Debugf("  Recv: %d Bytes", counter.BytesRecv)
				utils.Log().Debugf("  Packets Sent: %d", counter.PacketsSent)
				utils.Log().Debugf("  Packets Recv: %d", counter.PacketsRecv)

				if !netHistory[counter.Name].firstRun {
					// 计算网络IO速率（b/s）
					speedSent := float64(counter.BytesSent-netHistory[counter.Name].prevSent) * 8
					speedRecv := float64(counter.BytesRecv-netHistory[counter.Name].prevRecv) * 8
					// 计算网络包速率（pps）
					speedPacketsSent := float64(counter.PacketsSent - netHistory[counter.Name].prevPacketsSent)
					speedPacketsRecv := float64(counter.PacketsRecv - netHistory[counter.Name].prevPacketsRecv)

					// update net info
					_updateNetInfo(counter.Name, speedSent, speedRecv, speedPacketsSent, speedPacketsRecv)
				}

				// update history value
				netHistory[counter.Name] = struct {
					prevSent, prevRecv, prevPacketsSent, prevPacketsRecv uint64
					firstRun                                             bool
				}{
					prevSent:        counter.BytesSent,
					prevRecv:        counter.BytesRecv,
					prevPacketsSent: counter.PacketsSent,
					prevPacketsRecv: counter.PacketsRecv,
					firstRun:        false,
				}
			}
			time.Sleep(time.Second)
		}
	}()
}

func _updateNetInfo(netIf string, speedSent, speedRecv, speedPacketsSent, speedPacketsRecv float64) {
	netMu.Lock()
	defer netMu.Unlock()
	gNetInfo[netIf] = struct {
		Sent        float64
		Recv        float64
		PacketsSent float64
		PacketsRecv float64
	}{
		Sent:        speedSent,
		Recv:        speedRecv,
		PacketsSent: speedPacketsSent,
		PacketsRecv: speedPacketsRecv,
	}
}

func _getNetInfo(netIf string) (speedSent, speedRecv, speedPacketsSent, speedPacketsRecv float64) {
	netMu.RLock()
	defer netMu.RUnlock()
	return gNetInfo[netIf].Sent, gNetInfo[netIf].Recv, gNetInfo[netIf].PacketsSent, gNetInfo[netIf].PacketsRecv
}

func showMetric() error {
	var lines []string
	lines = append(lines, "=== System Monitor ===")

	if enableCpu {
		lines = append(lines, fmt.Sprintf("CPU: %.2f%%", gCpuPercent))
		// lines = append(lines, fmt.Sprintf("CPU: %.2f%% (Cores: physical[%d] / logic[%d])", gCpuPercent, gCpuPhysicalNum, gCpuLogicNum))
	}
	if enableMem {
		lines = append(lines, fmt.Sprintf("Memory: %dGB/%dGB (%.2f%%)",
			gMemUsed/1024/1024/1024,
			gMemTotal/1024/1024/1024,
			gMemPercent))
	}
	if enableNet {
		// 网络IO速率（Kb/s）当速率超过1024Kb/s时，单位转换为Mb/s，发送和接受分开计算
		// 遍历gNetInfo
		for _, netIf := range gNetNames {
			// 过滤出非回环接口
			if netIf == "lo" {
				continue
			}
			lines = append(lines, fmt.Sprintf("Interface %s:", netIf))
			// 计算网络IO速率（Kb/s）
			speedSent, speedRecv, speedPacketsSent, speedPacketsRecv := _getNetInfo(netIf)
			netSendMsg := fmt.Sprintf("%.2f %s", speedSent/1024, "Kb/s")
			netRecvMsg := fmt.Sprintf("%.2f %s", speedRecv/1024, "Kb/s")
			if speedSent/1024 > 1024 {
				netSendMsg = fmt.Sprintf("%.2f %s", speedSent/1024/1024, "Mb/s")
			}
			if speedRecv/1024 > 1024 {
				netRecvMsg = fmt.Sprintf("%.2f %s", speedRecv/1024/1024, "Mb/s")
			}
			lines = append(lines, fmt.Sprintf("  Network Sent: %s  Recv: %s.",
				netSendMsg,
				netRecvMsg))
			// 计算网络包速率（pps）
			lines = append(lines, fmt.Sprintf("  Packets Sent: %.2f pps  Recv: %.2f pps.",
				speedPacketsSent, speedPacketsRecv))
		}
	}

	lines = append(lines, "Press Ctrl+C to exit")

	if len(lines) == 2 { // 仅标题 + exit
		fmt.Println("no metric to monitor. -h for help")
		return nil
	}

	// 清屏后再输出（可选）
	fmt.Print("\033[H\033[2J")

	// 打印所有行
	for _, l := range lines {
		fmt.Println(l)
	}
	return nil
}

// show metric and send data to influxdb
func handleMetric() error {
	gTicker = utils.NewRepeatingTimer(interval, showMetric)
	gTicker.Start()
	return nil
}

// stop handle metric
func stopHandleMetric() error {
	if gTicker != nil {
		gTicker.Stop()
		gTicker = nil
	}
	return nil
}
