package main

import (
	"fmt"
	"net"
	"net/netip"
)

var _ net.Listener = &ipfilterListener{}

func IPFilterListener(l net.Listener, ipBlocks []netip.Prefix) net.Listener {
	return &ipfilterListener{
		Listener:  l,
		allowList: ipBlocks,
	}
}

type ipfilterListener struct {
	net.Listener
	allowList []netip.Prefix
}

func (i *ipfilterListener) Accept() (net.Conn, error) {
	conn, err := i.Listener.Accept()
	if err != nil {
		return conn, err
	}
	remoteIP, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return conn, err
	}
	ip, err := netip.ParseAddr(remoteIP)
	if err != nil {
		return conn, err
	}
	for _, prefix := range i.allowList {
		if prefix.Contains(ip) {
			return conn, nil
		}
	}
	return nil, fmt.Errorf("connection from ip %s not allowed", ip.String())
}
