package netlib

import "net"

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func NetworkContainsIP(network string, ip string) (bool, error) {
	_, ipv4Net, err := net.ParseCIDR(network)
	if err != nil {
		return false, err
	}
	ipo := net.ParseIP(ip)
	return ipv4Net.Contains(ipo), nil
}
