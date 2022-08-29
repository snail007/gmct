package tool

import "net"

func getLocalIP() (ips []string) {
	ifs, _ := net.Interfaces()
	for _, v := range ifs {
		addrs, err := v.Addrs()
		if err != nil {
			continue
		}
		for _, vv := range addrs {
			ip, _, err := net.ParseCIDR(vv.String())
			if err != nil {
				continue
			}
			if ip.To4() == nil || ip.IsLoopback() {
				continue
			}
			ips = append(ips, ip.String())
		}
	}
	return
}
