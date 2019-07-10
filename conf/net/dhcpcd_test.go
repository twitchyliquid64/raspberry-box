package net

import (
	"net"
	"testing"
)

func TestDHCPCD(t *testing.T) {
	tcs := []struct {
		name string
		inp  DHCPClientConf
		out  string
	}{
		{
			name: "empty",
			out:  "",
		},
		{
			name: "eth0 dhcp",
			inp: DHCPClientConf{
				Sections: []DHCPClientProfile{
					{
						InterfaceName: "eth0",
					},
				},
			},
			out: "interface eth0\ndhcp\n",
		},
		{
			name: "eth0 static",
			inp: DHCPClientConf{
				Sections: []DHCPClientProfile{
					{
						InterfaceName: "eth0",
						Mode:          ModeStatic,
						Static: struct {
							Network     net.IPNet
							Broadcast   net.IP
							IPv6Address net.IP
							Routers     []net.IP
							DNS         []string
						}{
							Network: net.IPNet{
								Mask: net.CIDRMask(24, 32),
								IP:   net.ParseIP("192.168.1.5"),
							},
							Routers: []net.IP{
								net.ParseIP("192.168.1.1"),
								net.ParseIP("192.168.1.11"),
							},
							DNS: []string{"8.8.8.8"},
						},
					},
				},
			},
			out: "interface eth0\nstatic ip_address=192.168.1.5/24\nstatic routers=192.168.1.1 192.168.1.11\nstatic domain_name_servers=8.8.8.8\n",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.out != tc.inp.String() {
				t.Errorf("out = %q, want %q", tc.inp.String(), tc.out)
			}
		})
	}
}
