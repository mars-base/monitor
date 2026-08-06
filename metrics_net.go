package main

import (
	"monitor/utils"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

func netMetric() {
	go func() {
		utils.Log().Debug("net metric started")

		netHistory := make(map[string]struct {
			prevSent, prevRecv, prevPacketsSent, prevPacketsRecv uint64
			firstRun                                             bool
		})

		// get net interface names
		netIfs, err := net.Interfaces()
		if err != nil {
			utils.Log().Errorf("failed to get net interfaces: %v", err)
			return
		}

		for _, netIf := range netIfs {
			utils.Log().Debugf("net interface: %s", netIf.Name)
			gNetInfo[netIf.Name] = NetInfo{}
			gNetNames = append(gNetNames, netIf.Name)
			netHistory[netIf.Name] = struct {
				prevSent, prevRecv, prevPacketsSent, prevPacketsRecv uint64
				firstRun                                             bool
			}{firstRun: true}
		}

		for {
			if stop {
				utils.Log().Debug("net metric stopped")
				return
			}

			netIO, err := net.IOCounters(true)
			if err != nil {
				utils.Log().Errorf("failed to get net io counters: %v", err)
				return
			}

			for _, counter := range netIO {
				if !netHistory[counter.Name].firstRun {
					speedSent := float64(counter.BytesSent-netHistory[counter.Name].prevSent) * 8
					speedRecv := float64(counter.BytesRecv-netHistory[counter.Name].prevRecv) * 8
					speedPacketsSent := float64(counter.PacketsSent - netHistory[counter.Name].prevPacketsSent)
					speedPacketsRecv := float64(counter.PacketsRecv - netHistory[counter.Name].prevPacketsRecv)

					updateNetInfo(counter.Name, speedSent, speedRecv, speedPacketsSent, speedPacketsRecv)
				}

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

func updateNetInfo(name string, sent, recv, pktSent, pktRecv float64) {
	netMu.Lock()
	defer netMu.Unlock()
	gNetInfo[name] = NetInfo{
		Sent:        sent,
		Recv:        recv,
		PacketsSent: pktSent,
		PacketsRecv: pktRecv,
	}
}

func getNetInfo(name string) NetInfo {
	netMu.RLock()
	defer netMu.RUnlock()
	return gNetInfo[name]
}
