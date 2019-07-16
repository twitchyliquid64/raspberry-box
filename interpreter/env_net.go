package interpreter

import (
	"crypto/sha256"
	"errors"
	"fmt"

	gonet "net"

	"github.com/twitchyliquid64/raspberry-box/conf/net"
	"go.starlark.net/starlark"
)

func netBuiltins(s *Script) starlark.StringDict {
	return starlark.StringDict{
		"DHCPProfile": starlark.NewBuiltin("DHCPProfile", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var name, interf, hostname starlark.String
			var leaseSeconds starlark.Int
			var clientID, persistent, rapidCommit, dns, ntp starlark.Bool
			if err := starlark.UnpackArgs("DHCPProfile", args, kwargs, "name?", &name, "interface", &interf, "hostname", &hostname,
				"client_id", &clientID, "persistent", &persistent, "rapid_commit", &rapidCommit, "dns", &dns, "request_ntp", &ntp,
				"lease_seconds", &leaseSeconds); err != nil {
				return starlark.None, err
			}
			p := &DHCPProfileProxy{
				Kind: "DHCP",
				Profile: &net.DHCPClientProfile{
					ProfileName:   string(name),
					InterfaceName: string(interf),
					Mode:          net.ModeDHCP,
				},
			}
			if i, ok := leaseSeconds.Int64(); ok {
				p.Profile.DHCP.LeaseSeconds = int(i)
			}
			p.Profile.DHCP.ClientID = bool(clientID)
			p.Profile.DHCP.Persistent = bool(persistent)
			p.Profile.DHCP.RapidCommit = bool(rapidCommit)
			p.Profile.DHCP.SetupDNS = bool(dns)
			p.Profile.DHCP.RequestNTP = bool(ntp)
			if string(hostname) != "" {
				p.Profile.DHCP.Hostname = string(hostname)
				p.Profile.DHCP.PresentHostname = true
			}
			return p, nil
		}),
		"StaticProfile": starlark.NewBuiltin("StaticProfile", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var name, interf, network, broadcast, ipv6 starlark.String
			var routers, dns *starlark.List
			if err := starlark.UnpackArgs("StaticProfile", args, kwargs, "name?", &name, "interface", &interf, "network", &network,
				"broadcast", &broadcast, "ipv6", &ipv6, "routers", &routers, "dns", &dns); err != nil {
				return starlark.None, err
			}
			p := &DHCPProfileProxy{
				Kind: "Static",
				Profile: &net.DHCPClientProfile{
					ProfileName:   string(name),
					InterfaceName: string(interf),
					Mode:          net.ModeStatic,
				},
			}

			ip, subnet, err := gonet.ParseCIDR(string(network))
			if err != nil {
				return nil, fmt.Errorf("parsing network: %v", err)
			}
			p.Profile.Static.Network = *subnet
			p.Profile.Static.Network.IP = ip
			p.Profile.Static.Broadcast = gonet.ParseIP(string(broadcast))
			p.Profile.Static.IPv6Address = gonet.ParseIP(string(ipv6))

			// Unpack routers.
			if routers != nil {
				for i := 0; i < routers.Len(); i++ {
					s, ok := routers.Index(i).(starlark.String)
					if !ok {
						return starlark.None, fmt.Errorf("routers[%d] is not a string", i)
					}
					p.Profile.Static.Routers = append(p.Profile.Static.Routers, gonet.ParseIP(string(s)))
				}
			}
			// Unpack DNS.
			if dns != nil {
				for i := 0; i < dns.Len(); i++ {
					s, ok := dns.Index(i).(starlark.String)
					if !ok {
						return starlark.None, fmt.Errorf("routers[%d] is not a string", i)
					}
					p.Profile.Static.DNS = append(p.Profile.Static.DNS, string(s))
				}
			}
			return p, nil
		}),
	}
}

// DHCPProfileProxy proxies access to a mounted filesystem.
type DHCPProfileProxy struct {
	Kind    string
	Profile *net.DHCPClientProfile
}

func (p *DHCPProfileProxy) String() string {
	return fmt.Sprintf("net.%sProfile{%p}", p.Kind, p)
}

// Type implements starlark.Value.
func (p *DHCPProfileProxy) Type() string {
	return fmt.Sprintf("net.%sProfile", p.Kind)
}

// Freeze implements starlark.Value.
func (p *DHCPProfileProxy) Freeze() {
}

// Truth implements starlark.Value.
func (p *DHCPProfileProxy) Truth() starlark.Bool {
	return starlark.Bool(true)
}

// Hash implements starlark.Value.
func (p *DHCPProfileProxy) Hash() (uint32, error) {
	h := sha256.Sum256([]byte(p.String()))
	return uint32(uint32(h[0]) + uint32(h[1])<<8 + uint32(h[2])<<16 + uint32(h[3])<<24), nil
}

// AttrNames implements starlark.Value.
func (p *DHCPProfileProxy) AttrNames() []string {
	return []string{"name", "interface", "hostname", "lease_seconds", "client_id", "persistent", "rapid_commit",
		"dns", "ntp", "set_name", "set_interface", "set_hostname", "set_lease_seconds", "set_client_id",
		"set_persistent", "set_rapid_commit", "set_dns", "set_ntp", "network", "set_network", "broadcast",
		"set_broadcast", "routers", "set_routers"}
}

func (p *DHCPProfileProxy) setName(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Profile.ProfileName = string(s)
	return starlark.None, nil
}

func (p *DHCPProfileProxy) setInterface(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Profile.InterfaceName = string(s)
	return starlark.None, nil
}

func (p *DHCPProfileProxy) setHostname(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Profile.DHCP.Hostname = string(s)
	p.Profile.DHCP.PresentHostname = p.Profile.DHCP.Hostname != ""
	return starlark.None, nil
}

func (p *DHCPProfileProxy) setLeaseSeconds(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	i, ok := args[0].(starlark.Int)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	in, ok := i.Int64()
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which cannot be represented as an integer")
	}
	p.Profile.DHCP.LeaseSeconds = int(in)
	return starlark.None, nil
}

func (p *DHCPProfileProxy) setClientID(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	b, ok := args[0].(starlark.Bool)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Profile.DHCP.ClientID = bool(b)
	return starlark.None, nil
}

func (p *DHCPProfileProxy) setPersistent(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	b, ok := args[0].(starlark.Bool)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Profile.DHCP.Persistent = bool(b)
	return starlark.None, nil
}

func (p *DHCPProfileProxy) setRapidCommit(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	b, ok := args[0].(starlark.Bool)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Profile.DHCP.RapidCommit = bool(b)
	return starlark.None, nil
}

func (p *DHCPProfileProxy) setDNS(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if p.Kind == "DHCP" {
		b, ok := args[0].(starlark.Bool)
		if !ok {
			return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
		}
		p.Profile.DHCP.SetupDNS = bool(b)
		return starlark.None, nil
	}

	dns, ok := args[0].(*starlark.List)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	var out []string
	for i := 0; i < dns.Len(); i++ {
		s, ok := dns.Index(i).(starlark.String)
		if !ok {
			return starlark.None, fmt.Errorf("dns[%d] is not a string", i)
		}
		out = append(out, string(s))
	}
	p.Profile.Static.DNS = out
	return starlark.None, nil
}

func (p *DHCPProfileProxy) setNTP(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	b, ok := args[0].(starlark.Bool)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Profile.DHCP.RequestNTP = bool(b)
	return starlark.None, nil
}

func (p *DHCPProfileProxy) setNetwork(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	ip, n, err := gonet.ParseCIDR(string(s))
	if err != nil {
		return starlark.None, err
	}
	p.Profile.Static.Network = *n
	p.Profile.Static.Network.IP = ip
	return starlark.None, nil
}

func (p *DHCPProfileProxy) setBroadcast(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Profile.Static.Broadcast = gonet.ParseIP(string(s))
	return starlark.None, nil
}

func (p *DHCPProfileProxy) setRouters(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	routers, ok := args[0].(*starlark.List)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	var out []gonet.IP
	for i := 0; i < routers.Len(); i++ {
		s, ok := routers.Index(i).(starlark.String)
		if !ok {
			return starlark.None, fmt.Errorf("routers[%d] is not a string", i)
		}
		out = append(out, gonet.ParseIP(string(s)))
	}
	p.Profile.Static.Routers = out
	return starlark.None, nil
}

// Attr implements starlark.Value.
func (p *DHCPProfileProxy) Attr(name string) (starlark.Value, error) {
	switch name {
	case "name":
		return starlark.String(p.Profile.ProfileName), nil
	case "set_name":
		return starlark.NewBuiltin("set_name", p.setName), nil
	case "interface":
		return starlark.String(p.Profile.InterfaceName), nil
	case "set_interface":
		return starlark.NewBuiltin("set_interface", p.setInterface), nil
	case "hostname":
		return starlark.String(p.Profile.DHCP.Hostname), nil
	case "set_hostname":
		return starlark.NewBuiltin("set_hostname", p.setHostname), nil
	case "lease_seconds":
		return starlark.MakeInt(p.Profile.DHCP.LeaseSeconds), nil
	case "set_lease_seconds":
		return starlark.NewBuiltin("set_lease_seconds", p.setLeaseSeconds), nil

	case "client_id":
		return starlark.Bool(p.Profile.DHCP.ClientID), nil
	case "set_client_id":
		return starlark.NewBuiltin("set_client_id", p.setClientID), nil
	case "persistent":
		return starlark.Bool(p.Profile.DHCP.Persistent), nil
	case "set_persistent":
		return starlark.NewBuiltin("set_persistent", p.setPersistent), nil
	case "rapid_commit":
		return starlark.Bool(p.Profile.DHCP.RapidCommit), nil
	case "set_rapid_commit":
		return starlark.NewBuiltin("set_rapid_commit", p.setRapidCommit), nil
	case "dns":
		if p.Kind == "DHCP" {
			return starlark.Bool(p.Profile.DHCP.SetupDNS), nil
		}
		var out []starlark.Value
		for _, r := range p.Profile.Static.DNS {
			out = append(out, starlark.String(r))
		}
		return starlark.NewList(out), nil
	case "set_dns":
		return starlark.NewBuiltin("set_dns", p.setDNS), nil
	case "ntp":
		return starlark.Bool(p.Profile.DHCP.RequestNTP), nil
	case "set_ntp":
		return starlark.NewBuiltin("set_ntp", p.setNTP), nil

	case "network":
		return starlark.String(p.Profile.Static.Network.String()), nil
	case "set_network":
		return starlark.NewBuiltin("set_network", p.setNetwork), nil
	case "broadcast":
		return starlark.String(p.Profile.Static.Broadcast.String()), nil
	case "set_broadcast":
		return starlark.NewBuiltin("set_broadcast", p.setBroadcast), nil
	case "routers":
		var out []starlark.Value
		for _, r := range p.Profile.Static.Routers {
			out = append(out, starlark.String(r.String()))
		}
		return starlark.NewList(out), nil
	case "set_routers":
		return starlark.NewBuiltin("set_routers", p.setRouters), nil
	}

	return nil, starlark.NoSuchAttrError(
		fmt.Sprintf("%s has no .%s attribute", p.Type(), name))
}

// SetField implements starlark.HasSetField.
func (p *DHCPProfileProxy) SetField(name string, val starlark.Value) error {
	switch name {
	case "routers":
		_, err := p.setRouters(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "dns":
		_, err := p.setDNS(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "name":
		_, err := p.setName(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "interface":
		_, err := p.setInterface(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "hostname":
		_, err := p.setHostname(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "broadcast":
		_, err := p.setBroadcast(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "network":
		_, err := p.setNetwork(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	}
	return errors.New("no such assignable field: " + name)
}
