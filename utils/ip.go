package utils

import (
	"net"
)

func GetIPAddrList() []string {
	var ips []string

	interfaces, _ := net.Interfaces()

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, _ := iface.Addrs()

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			if ip.To4() == nil {
				continue
			}

			ipStr := ip.String()
			ips = append(ips, ipStr)
		}
	}

	return ips
}
