// Package net generates config files for network management daemons.
package net

import (
	"fmt"
	"net"
	"strings"
)

// DHCPClientConf encapsulates configuration of the dhcp client.
type DHCPClientConf struct {
	ControlGroup string // Group to set ownership of control socket.
	Sections     []DHCPClientProfile
}

// DHCPMode controls how an interface behaves.
type DHCPMode uint8

// Valid DHCPMode values.
const (
	ModeDHCP DHCPMode = iota
	ModeStatic
)

// DHCPClientProfile describes a unit of dhcpcd configuration. Only one
// of InterfaceName or ProfileName should be set.
type DHCPClientProfile struct {
	ProfileName   string
	InterfaceName string
	Mode          DHCPMode

	DHCP struct {
		LeaseSeconds    int
		ClientID        bool
		PresentHostname bool
		Hostname        string
	}

	Static struct {
		Network     net.IPNet
		Broadcast   net.IP
		IPv6Address net.IP
		Routers     []net.IP
		DNS         []string
	}
}

// String returns a representation ready to use in a config file.
func (p *DHCPClientProfile) String() string {
	var out strings.Builder

	if p.ProfileName != "" {
		out.WriteString("profile " + p.ProfileName + "\n")
	} else if p.InterfaceName != "" {
		out.WriteString("interface " + p.InterfaceName + "\n")
	}

	switch p.Mode {
	case ModeDHCP:
		out.WriteString("dhcp\n")
		if p.DHCP.PresentHostname {
			if p.DHCP.Hostname != "" {
				out.WriteString("hostname " + p.DHCP.Hostname + "\n")
			} else {
				out.WriteString("hostname\n")
			}
		}
		if p.DHCP.ClientID {
			out.WriteString("clientid\n")
		}
		if p.DHCP.LeaseSeconds > 0 {
			out.WriteString(fmt.Sprintf("leasetime %d\n", p.DHCP.LeaseSeconds))
		}

	case ModeStatic:
		out.WriteString(fmt.Sprintf("static ip_address=%s\n", p.Static.Network.String()))
		if len(p.Static.Broadcast) > 0 {
			out.WriteString(fmt.Sprintf("static broadcast_address=%s\n", p.Static.Broadcast.String()))
		}
		if len(p.Static.IPv6Address) > 0 {
			out.WriteString(fmt.Sprintf("static ip6_address=%s\n", p.Static.IPv6Address.String()))
		}
		if len(p.Static.Routers) > 0 {
			r := make([]string, len(p.Static.Routers))
			for i, router := range p.Static.Routers {
				r[i] = router.String()
			}
			out.WriteString(fmt.Sprintf("static routers=%s\n", strings.Join(r, " ")))
		}
		if len(p.Static.DNS) > 0 {
			out.WriteString(fmt.Sprintf("static domain_name_servers=%s\n", strings.Join(p.Static.DNS, " ")))
		}
	}

	return out.String()
}

// String returns a representation ready to use in a config file.
func (c *DHCPClientConf) String() string {
	var out strings.Builder

	if c.ControlGroup != "" {
		out.WriteString("controlgroup " + c.ControlGroup + "\n")
	}

	for i, section := range c.Sections {
		out.WriteString(section.String())
		if i < len(c.Sections)-1 {
			out.WriteString("\n")
		}
	}

	return out.String()
}
