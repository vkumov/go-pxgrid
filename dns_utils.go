package gopxgrid

import (
	"net"
	"strings"
)

func tryParseUDP(host string) (*net.UDPAddr, error) {
	u, err := net.ResolveUDPAddr("udp", host)
	if err != nil && strings.Contains(err.Error(), "missing port in address") {
		if strings.Contains(host, "]") {
			host = host + ":53"
		} else {
			host = net.JoinHostPort(host, "53")
		}
	} else if u == nil || u.IP == nil {
		return nil, err
	}

	u, err = net.ResolveUDPAddr("udp", host)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func tryParseIP(host string) (*net.UDPAddr, error) {
	ip, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil, err
	}

	return &net.UDPAddr{
		IP:   ip.IP,
		Port: 53,
		Zone: ip.Zone,
	}, nil
}

func ParseDNSHost(host string) (*net.UDPAddr, error) {
	u, err := tryParseUDP(host)
	if err == nil {
		return u, nil
	}

	return tryParseIP(host)
}
