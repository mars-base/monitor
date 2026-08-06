package main

import (
	"fmt"
	"monitor/utils"
	"os"

	"golang.org/x/term"
)

func showMetric() error {
	if !enableCpu && !enableMem && !enableNet && !enableDisk {
		fmt.Println("no metric to monitor. -h for help")
		return nil
	}

	lines := []string{centerText("=== System Monitor ===")}

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
		lines = appendNetLines(lines)
	}
	if enableDisk {
		lines = appendDiskLines(lines)
	}

	lines = append(lines, "Press Ctrl+C to exit")

	utils.DrawTabTerm(lines)
	return nil
}

func appendNetLines(lines []string) []string {
	const (
		kb = 1024
		mb = kb * 1024
	)

	for _, name := range gNetNames {
		if name == "lo" {
			continue
		}
		if interfaceName != "" && name != interfaceName {
			continue
		}

		info := getNetInfo(name)

		sendMsg := formatBandwidth(info.Sent, "b/s", kb, "Kb/s", mb, "Mb/s")
		recvMsg := formatBandwidth(info.Recv, "b/s", kb, "Kb/s", mb, "Mb/s")

		lines = append(lines, fmt.Sprintf("Interface %s:\t%s\t%s", name, "", ""))
		lines = append(lines, fmt.Sprintf("  Network\tSent: %s\tRecv: %s.", sendMsg, recvMsg))
		lines = append(lines, fmt.Sprintf("  Packets\tSent: %.2f pps\tRecv: %.2f pps.", info.PacketsSent, info.PacketsRecv))
	}
	return lines
}

func appendDiskLines(lines []string) []string {
	const (
		kb = 1024
		mb = kb * 1024
	)

	for _, name := range gDiskNames {
		info := getDiskInfo(name)

		readMsg := formatBandwidth(float64(info.Read), "B/s", kb, "KB/s", mb, "MB/s")
		writeMsg := formatBandwidth(float64(info.Write), "B/s", kb, "KB/s", mb, "MB/s")

		lines = append(lines, fmt.Sprintf("Disk %s:\t%s\t%s", name, "", ""))
		lines = append(lines, fmt.Sprintf("  \tRead: %s\tWrite: %s.", readMsg, writeMsg))
	}
	return lines
}

// formatBandwidth formats a value with auto-scaling unit
func formatBandwidth(val float64, unitLow string, mid int, unitMid string, high int, unitHigh string) string {
	if int(val) > high {
		return fmt.Sprintf("%.2f %s", val/float64(high), unitHigh)
	}
	if int(val) > mid {
		return fmt.Sprintf("%.2f %s", val/float64(mid), unitMid)
	}
	return fmt.Sprintf("%.2f %s", val, unitLow)
}

func handleMetric() error {
	gTicker = utils.NewRepeatingTimer(interval, showMetric)
	gTicker.Start()
	return nil
}

func stopHandleMetric() error {
	if gTicker != nil {
		gTicker.Stop()
		gTicker = nil
	}
	return nil
}

func centerText(text string) string {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		width = 80
	}
	padding := max(0, (width-len(text))/2)
	return fmt.Sprintf("%*s%s", padding, "", text)
}
