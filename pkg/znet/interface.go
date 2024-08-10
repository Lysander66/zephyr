package znet

import (
	"net"
)

const localhost = "127.0.0.1"

func OutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return localhost
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func IntranetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return localhost
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			if ipNet.IP.IsGlobalUnicast() {
				return ipNet.IP.String()
			}
		}
	}
	return localhost
}

func PublicIPs() (ips []string) {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			panic(err)
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
				if isPublicIP(ipNet.IP) {
					ips = append(ips, ipNet.IP.String())
				}
			}
		}
	}

	return
}

func isPublicIP(ip net.IP) bool {
	var privateIPBlocks = []*net.IPNet{
		// 10.0.0.0/8
		{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		// 172.16.0.0/12
		{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		// 192.168.0.0/16
		{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return false
		}
	}

	return true
}
