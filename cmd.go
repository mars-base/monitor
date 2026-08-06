package main

import (
	"fmt"
	"monitor/utils"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// global variable
var (
	stop    bool = false
	gTicker *utils.RepeatingTimer

	enableCpu     bool
	enableMem     bool
	enableNet     bool
	enableDisk    bool
	enableAll     bool
	interval      int
	interfaceName string // 网络接口名称，默认所有接口

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

var (
	diskMu    sync.RWMutex
	gDiskInfo = make(map[string]struct {
		Read  uint64 // 读取字节数（B/s）
		Write uint64 // 写入字节数（B/s）
	})
	gDiskNames = []string{}
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
		utils.Log().Debug(AppName, AppVersion)
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
		fmt.Println("enable net:", enableDisk)
		fmt.Println("enable all:", enableAll)
		fmt.Printf("interval: %d seconds\n", interval)
		fmt.Println("interface name:", interfaceName)
	},
}

// CmdEntryPoint is the entry point of cobra command,
// it should be called in main.go
func CmdEntryPoint() {
	// add flag
	rootCmd.PersistentFlags().BoolVarP(&enableCpu, "cpu", "c", false, "enable cpu metric")
	rootCmd.PersistentFlags().BoolVarP(&enableMem, "mem", "m", false, "enable mem metric")
	rootCmd.PersistentFlags().BoolVarP(&enableNet, "net", "n", false, "enable net metric")
	rootCmd.PersistentFlags().BoolVarP(&enableDisk, "disk", "d", false, "enable disk metric")
	// -a, --all 显示cpu/mem/网络接口的统计信息
	rootCmd.PersistentFlags().BoolVarP(&enableAll, "all", "a", false, "enable all metrics")
	rootCmd.PersistentFlags().IntVarP(&interval, "interval", "i", 2, "interval in seconds")
	// --if 指定监控的网络接口名称，默认所有接口
	rootCmd.PersistentFlags().StringVarP(&interfaceName, "if", "", "", "network interface name")

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
		if enableCpu || enableMem || enableNet || enableDisk {
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
			if enableDisk {
				utils.Log().Debug("start disk monitor module")
				diskMetric()
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

func isPhysicalDiskOnLinux(name string) bool {
	// 先排除明显虚拟设备
	if isVirtualDevice(name) {
		return false
	}

	// 再通过 /sys/block/xxx/device 确认是物理设备
	_, err := os.Stat("/sys/block/" + name + "/device")
	return err == nil
}

func isVirtualDevice(name string) bool {
	if name == "" {
		return true
	}
	prefixes := []string{"dm-", "loop", "ram", "sr", "fd", "md", "zram"}
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// isPhysicalDisk 跨平台判断（目前仅 Linux 有精准方法）
func isPhysicalDisk(name string) bool {
	if runtime.GOOS == "linux" {
		return isPhysicalDiskOnLinux(name)
	}
	// Windows/macOS: 暂按名称粗略过滤（可根据需求增强）
	return true // 或根据名称规则过滤
}

func diskMetric() {
	go func() {
		utils.Log().Debug("disk metric started")

		diskHistory := make(map[string]struct {
			prevRead, prevWrite uint64
			firstRun            bool
		})

		// 获取磁盘设备
		diskIfs, err := disk.IOCounters()
		if err != nil {
			utils.Log().Errorf("failed to get disk io counters: %v", err)
			return
		}
		// 初始化磁盘历史记录
		for name, counter := range diskIfs {
			if isPhysicalDisk(name) {
				// 记录物理磁盘名称
				gDiskNames = append(gDiskNames, name)
				diskHistory[name] = struct {
					prevRead, prevWrite uint64
					firstRun            bool
				}{
					prevRead:  counter.ReadBytes,
					prevWrite: counter.WriteBytes,
					firstRun:  true,
				}
			}
		}

		time.Sleep(time.Second)

		for {
			if stop {
				utils.Log().Debug("disk metric stopped")
				return
			}
			utils.Log().Debug("collect disk info")
			diskIO, err := disk.IOCounters()
			if err != nil {
				utils.Log().Errorf("failed to get disk io counters: %v", err)
				return
			}
			// 遍历每个磁盘接口
			for name, counter := range diskIO {
				// 记录物理磁盘统计信息
				if !isPhysicalDisk(name) {
					continue
				}
				utils.Log().Debugf("Disk interface: %s", name)
				utils.Log().Debugf("  Read Bytes: %d", counter.ReadBytes)
				utils.Log().Debugf("  Write Bytes: %d", counter.WriteBytes)

				// update disk info
				if !diskHistory[name].firstRun {
					_updateDiskInfo(name, counter.ReadBytes-diskHistory[name].prevRead, counter.WriteBytes-diskHistory[name].prevWrite)
				}

				// update history value
				diskHistory[name] = struct {
					prevRead, prevWrite uint64
					firstRun            bool
				}{
					prevRead:  counter.ReadBytes,
					prevWrite: counter.WriteBytes,
					firstRun:  false,
				}
			}
			time.Sleep(time.Second)
		}
	}()
}

func _updateDiskInfo(diskIf string, read, write uint64) {
	diskMu.Lock()
	defer diskMu.Unlock()
	gDiskInfo[diskIf] = struct {
		Read  uint64 // 读取字节数（B/s）
		Write uint64 // 写入字节数（B/s）
	}{
		Read:  read,
		Write: write,
	}
}

func _getDiskInfo(diskIf string) (read, write uint64) {
	diskMu.RLock()
	defer diskMu.RUnlock()
	return gDiskInfo[diskIf].Read, gDiskInfo[diskIf].Write
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
	if !enableCpu && !enableMem && !enableNet && !enableDisk {
		fmt.Println("no metric to monitor. -h for help")
		return nil
	}

	lines := []string{}
	lines = append(lines, centerText("=== System Monitor ==="))
	if enableCpu {
		lines = append(lines, fmt.Sprintf("CPU:\t%d/%d\t%.2f%%", gCpuPhysicalNum, gCpuLogicNum, gCpuPercent))
	}
	if enableMem {
		lines = append(lines, fmt.Sprintf("Memory:\t%dGB/%dGB\t%.2f%%",
			gMemUsed/1024/1024/1024,
			gMemTotal/1024/1024/1024,
			gMemPercent))
	}
	if enableNet {
		for _, netIf := range gNetNames {
			// exclude loopback interface
			if netIf == "lo" {
				continue
			}
			// just filter spec net if
			if interfaceName != "" && netIf != interfaceName {
				continue
			}

			speedSent, speedRecv, speedPacketsSent, speedPacketsRecv := _getNetInfo(netIf)

			const (
				b  = 1
				kb = 1024
				mb = kb * 1024
			)

			netSendMsg := fmt.Sprintf("%.2f b/s", speedSent)
			netRecvMsg := fmt.Sprintf("%.2f b/s", speedRecv)

			if speedSent > kb {
				netSendMsg = fmt.Sprintf("%.2f Kb/s", speedSent/kb)
			}
			if speedRecv > kb {
				netRecvMsg = fmt.Sprintf("%.2f Kb/s", speedRecv/kb)
			}
			if speedSent > mb {
				netSendMsg = fmt.Sprintf("%.2f Mb/s", speedSent/mb)
			}
			if speedRecv > mb {
				netRecvMsg = fmt.Sprintf("%.2f Mb/s", speedRecv/mb)
			}

			lines = append(lines, fmt.Sprintf("Interface %s:\t%s\t%s", netIf, "", ""))
			lines = append(lines, fmt.Sprintf("  Network\tSent: %s\tRecv: %s.", netSendMsg, netRecvMsg))
			lines = append(lines, fmt.Sprintf("  Packets\tSent: %.2f pps\tRecv: %.2f pps.", speedPacketsSent, speedPacketsRecv))
		}
	}

	if enableDisk {
		for _, diskIf := range gDiskNames {
			read, write := _getDiskInfo(diskIf)
			const (
				b  = 1
				kb = 1024
				mb = kb * 1024
			)

			diskReadMsg := fmt.Sprintf("%.2f B/s", float64(read))
			diskWriteMsg := fmt.Sprintf("%.2f B/s", float64(write))

			if read > kb {
				diskReadMsg = fmt.Sprintf("%.2f KB/s", float64(read)/kb)
			}
			if write > kb {
				diskWriteMsg = fmt.Sprintf("%.2f KB/s", float64(write)/kb)
			}
			if read > mb {
				diskReadMsg = fmt.Sprintf("%.2f MB/s", float64(read)/mb)
			}
			if write > mb {
				diskWriteMsg = fmt.Sprintf("%.2f MB/s", float64(write)/mb)
			}

			lines = append(lines, fmt.Sprintf("Disk %s:\t%s\t%s", diskIf, "", ""))
			lines = append(lines, fmt.Sprintf("  \tRead: %s\tWrite: %s.", diskReadMsg, diskWriteMsg))
		}
	}

	lines = append(lines, "Press Ctrl+C to exit")

	// draw terminal
	utils.DrawTabTerm(lines)

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
