package main

import (
	"monitor/utils"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

func diskMetric() {
	go func() {
		utils.Log().Debug("disk metric started")

		diskHistory := make(map[string]struct {
			prevRead, prevWrite uint64
			firstRun            bool
		})

		// discover physical disks
		diskIfs, err := disk.IOCounters()
		if err != nil {
			utils.Log().Errorf("failed to get disk io counters: %v", err)
			return
		}
		for name, counter := range diskIfs {
			if isPhysicalDisk(name) {
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

			diskIO, err := disk.IOCounters()
			if err != nil {
				utils.Log().Errorf("failed to get disk io counters: %v", err)
				return
			}

			for name, counter := range diskIO {
				if !isPhysicalDisk(name) {
					continue
				}

				if !diskHistory[name].firstRun {
					read := counter.ReadBytes - diskHistory[name].prevRead
					write := counter.WriteBytes - diskHistory[name].prevWrite
					updateDiskInfo(name, read, write)
				}

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

func updateDiskInfo(name string, read, write uint64) {
	diskMu.Lock()
	defer diskMu.Unlock()
	gDiskInfo[name] = DiskInfo{Read: read, Write: write}
}

func getDiskInfo(name string) DiskInfo {
	diskMu.RLock()
	defer diskMu.RUnlock()
	return gDiskInfo[name]
}

// isPhysicalDisk checks if a disk device is physical (not virtual)
func isPhysicalDisk(name string) bool {
	if runtime.GOOS == "linux" {
		return isPhysicalDiskLinux(name)
	}
	return true
}

func isPhysicalDiskLinux(name string) bool {
	if isVirtualDevice(name) {
		return false
	}
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
