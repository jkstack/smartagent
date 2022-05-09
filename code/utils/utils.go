package utils

import "net"

type ipIntf interface {
	LocalAddr() net.Addr
}

// GetIP get ip address
func GetIP(c ipIntf) net.IP {
	var ret net.IP
	host, _, err := net.SplitHostPort(c.LocalAddr().String())
	if err == nil {
		ret = net.ParseIP(host)
	} else {
		ret = net.ParseIP(c.LocalAddr().String())
	}
	return ret
}

// GetMac get mac address
func GetMac(ip net.IP) string {
	intfs, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, intf := range intfs {
		addrs, err := intf.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			intfIP, _, _ := net.ParseCIDR(addr.String())
			if intfIP.Equal(ip) {
				return intf.HardwareAddr.String()
			}
		}
	}
	return ""
}
