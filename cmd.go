package main

import (
	"bytes"
	"fmt"
	"monitor/utils"
	"os"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// global variable
var (
	stop    bool = false
	gTicker *utils.RepeatingTimer

	enableCpu bool
	enableMem bool
	enableNet bool
	enableAll bool
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
		fmt.Println("enable cpu:", enableCpu)
		fmt.Println("enable mem:", enableMem)
		fmt.Println("enable net:", enableNet)
		fmt.Println("enable all:", enableAll)
		fmt.Printf("interval: %d seconds\n", interval)
	},
}

// CmdEntryPoint is the entry point of cobra command,
// it should be called in main.go
func CmdEntryPoint() {
	// add flag
	rootCmd.PersistentFlags().BoolVarP(&enableCpu, "cpu", "c", false, "enable cpu metric")
	rootCmd.PersistentFlags().BoolVarP(&enableMem, "mem", "m", false, "enable mem metric")
	rootCmd.PersistentFlags().BoolVarP(&enableNet, "net", "n", false, "enable net metric")
	// -a, --all 显示cpu/mem/网络接口的统计信息
	rootCmd.PersistentFlags().BoolVarP(&enableAll, "all", "a", true, "enable all metrics")
	rootCmd.PersistentFlags().IntVarP(&interval, "interval", "i", 2, "interval in seconds")

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
		utils.Log().Debug("enable all:", enableAll)
		utils.Log().Debug("interval seconds:", interval)

		if enableAll {
			enableCpu = true
			enableMem = true
			enableNet = true
		}
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
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, centerText("=== System Monitor ==="))

	if enableCpu {
		fmt.Fprintf(w, "CPU:\t%d/%d\t%.2f%%\n", gCpuPhysicalNum, gCpuLogicNum, gCpuPercent)
	}
	if enableMem {
		fmt.Fprintf(w, "Memory:\t%dGB/%dGB\t%.2f%%\n",
			gMemUsed/1024/1024/1024,
			gMemTotal/1024/1024/1024,
			gMemPercent)
	}
	if enableNet {
		for _, netIf := range gNetNames {
			if netIf == "lo" {
				continue
			}
			speedSent, speedRecv, speedPacketsSent, speedPacketsRecv := _getNetInfo(netIf)

			netSendMsg := fmt.Sprintf("%.2f Kb/s", speedSent/1024)
			netRecvMsg := fmt.Sprintf("%.2f Kb/s", speedRecv/1024)
			if speedSent/1024 > 1024 {
				netSendMsg = fmt.Sprintf("%.2f Mb/s", speedSent/1024/1024)
			}
			if speedRecv/1024 > 1024 {
				netRecvMsg = fmt.Sprintf("%.2f Mb/s", speedRecv/1024/1024)
			}

			fmt.Fprintf(w, "Interface %s:\t%s\t%s\n", netIf, "", "")
			fmt.Fprintf(w, "  Network Sent:\t%s\tRecv: %s.\n", netSendMsg, netRecvMsg)
			fmt.Fprintf(w, "  Packets Sent:\t%.2f pps\tRecv: %.2f pps.\n", speedPacketsSent, speedPacketsRecv)
		}
	}

	fmt.Fprintln(w, "Press Ctrl+C to exit")
	w.Flush()

	if buf.Len() <= len("=== System Monitor ===\nPress Ctrl+C to exit\n") {
		fmt.Println("no metric to monitor. -h for help")
		return nil
	}

	// 清屏后再输出
	fmt.Print("\033[H\033[2J")
	fmt.Print(buf.String())

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

// 居中打印标题
func centerText(text string) string {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		width = 80 // 默认终端宽度
	}
	padding := (width - len(text)) / 2
	if padding < 0 {
		padding = 0
	}
	return fmt.Sprintf("%*s%s", padding, "", text)
}
