package interpreter

import (
	"crypto/sha256"
	"errors"
	"fmt"

	gonet "net"

	"github.com/twitchyliquid64/raspberry-box/conf/net"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func netBuiltins(s *Script) starlark.StringDict {
	return starlark.StringDict{
		"wifi": starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
			"mode_client": starlark.MakeUint(uint(net.ModeClient)),
			"mode_adhoc":  starlark.MakeUint(uint(net.ModeAdhoc)),
			"mode_ap":     starlark.MakeUint(uint(net.ModeAP)),
			"Network": starlark.NewBuiltin("Network", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
				var ssid, psk starlark.String
				var mode starlark.Int
				var disabled starlark.Bool
				if err := starlark.UnpackArgs("Network", args, kwargs, "mode?", &mode, "ssid", &ssid,
					"psk", &psk, "disabled", &disabled); err != nil {
					return starlark.None, err
				}
				m, ok := mode.Uint64()
				if !ok {
					return starlark.None, errors.New("mode must be an unsigned integer")
				}
				p := &WifiNetworkProxy{
					Conf: &net.WPASupplicantNetwork{
						Mode:     net.WPASupplicantMode(m),
						SSID:     string(ssid),
						PSK:      string(psk),
						Disabled: bool(disabled),
					},
				}
				return p, nil
			}),
			"SupplicantConfig": starlark.NewBuiltin("SupplicantConfig", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
				var ctrlInterface, ctrlInterfaceGroup starlark.String
				var countryCode, deviceName starlark.String
				var allowUpdateConfig starlark.Bool
				var networks *starlark.List
				if err := starlark.UnpackArgs("SupplicantConfig", args, kwargs, "control_interface?", &ctrlInterface, "control_interface_group", &ctrlInterfaceGroup,
					"allow_update_config", &allowUpdateConfig, "country_code", &countryCode, "device_name", &deviceName, "networks", &networks); err != nil {
					return starlark.None, err
				}
				p := &WifiConfigurationProxy{
					Conf: &net.WPASupplicantConfig{
						CtrlInterface:      string(ctrlInterface),
						CtrlInterfaceGroup: string(ctrlInterfaceGroup),
						CountryCode:        string(countryCode),
						DeviceName:         string(deviceName),
						AllowUpdateConfig:  bool(allowUpdateConfig),
					},
				}
				// Unpack networks.
				if networks != nil {
					for i := 0; i < networks.Len(); i++ {
						s, ok := networks.Index(i).(*WifiNetworkProxy)
						if !ok {
							return starlark.None, fmt.Errorf("networks[%d] is not a net.WifiNetwork", i)
						}
						p.Conf.Networks = append(p.Conf.Networks, *s.Conf)
					}
				}
				return p, nil
			}),
		}),
		"DHCPClient": starlark.NewBuiltin("DHCPClient", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var controlgroup starlark.String
			var profiles *starlark.List
			if err := starlark.UnpackArgs("DHCPClient", args, kwargs, "control_group?", &controlgroup, "profiles", &profiles); err != nil {
				return starlark.None, err
			}
			p := &DHCPClientConfProxy{
				Conf: &net.DHCPClientConf{
					ControlGroup: string(controlgroup),
				},
			}
			// Unpack profiles.
			if profiles != nil {
				for i := 0; i < profiles.Len(); i++ {
					s, ok := profiles.Index(i).(*DHCPProfileProxy)
					if !ok {
						return starlark.None, fmt.Errorf("profiles[%d] is not a net.DHCPProfile or net.StaticProfile", i)
					}
					p.Conf.Sections = append(p.Conf.Sections, *s.Profile)
				}
			}
			return p, nil
		}),
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

// DHCPProfileProxy proxies access to DHCP profile structure.
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

// DHCPClientConfProxy proxies access to DHCPClientConf structure.
type DHCPClientConfProxy struct {
	Conf *net.DHCPClientConf
}

func (p *DHCPClientConfProxy) String() string {
	return p.Conf.String()
}

// Type implements starlark.Value.
func (p *DHCPClientConfProxy) Type() string {
	return fmt.Sprintf("net.DHCPClient")
}

// Freeze implements starlark.Value.
func (p *DHCPClientConfProxy) Freeze() {
}

// Truth implements starlark.Value.
func (p *DHCPClientConfProxy) Truth() starlark.Bool {
	return starlark.Bool(true)
}

// Hash implements starlark.Value.
func (p *DHCPClientConfProxy) Hash() (uint32, error) {
	h := sha256.Sum256([]byte(p.String()))
	return uint32(uint32(h[0]) + uint32(h[1])<<8 + uint32(h[2])<<16 + uint32(h[3])<<24), nil
}

// AttrNames implements starlark.Value.
func (p *DHCPClientConfProxy) AttrNames() []string {
	return []string{"control_group", "set_control_group", "profiles", "set_profiles"}
}

// Attr implements starlark.Value.
func (p *DHCPClientConfProxy) Attr(name string) (starlark.Value, error) {
	switch name {
	case "control_group":
		return starlark.String(p.Conf.ControlGroup), nil
	case "set_control_group":
		return starlark.NewBuiltin("set_control_group", p.setControlGroup), nil

	case "profiles":
		var out []starlark.Value
		for _, s := range p.Conf.Sections {
			out = append(out, &DHCPProfileProxy{Profile: &s})
		}
		return starlark.NewList(out), nil
	case "set_profiles":
		return starlark.NewBuiltin("set_profiles", p.setProfiles), nil

	}

	return nil, starlark.NoSuchAttrError(
		fmt.Sprintf("%s has no .%s attribute", p.Type(), name))
}

func (p *DHCPClientConfProxy) setControlGroup(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.ControlGroup = string(s)
	return starlark.None, nil
}

func (p *DHCPClientConfProxy) setProfiles(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	profiles, ok := args[0].(*starlark.List)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	var out []net.DHCPClientProfile
	for i := 0; i < profiles.Len(); i++ {
		prof, ok := profiles.Index(i).(*DHCPProfileProxy)
		if !ok {
			return starlark.None, fmt.Errorf("profiles[%d] is not a StaticProfile or DHCPProfile", i)
		}
		out = append(out, *prof.Profile)
	}
	p.Conf.Sections = out
	return starlark.None, nil
}

// SetField implements starlark.HasSetField.
func (p *DHCPClientConfProxy) SetField(name string, val starlark.Value) error {
	switch name {
	case "profiles":
		_, err := p.setProfiles(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "control_group":
		_, err := p.setControlGroup(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	}
	return errors.New("no such assignable field: " + name)
}

// WifiNetworkProxy proxies access to a net.WPASupplicantNetwork structure.
type WifiNetworkProxy struct {
	Conf *net.WPASupplicantNetwork
}

func (p *WifiNetworkProxy) String() string {
	return p.Conf.String()
}

// Type implements starlark.Value.
func (p *WifiNetworkProxy) Type() string {
	return fmt.Sprintf("net.WifiNetwork")
}

// Freeze implements starlark.Value.
func (p *WifiNetworkProxy) Freeze() {
}

// Truth implements starlark.Value.
func (p *WifiNetworkProxy) Truth() starlark.Bool {
	return starlark.Bool(true)
}

// Hash implements starlark.Value.
func (p *WifiNetworkProxy) Hash() (uint32, error) {
	h := sha256.Sum256([]byte(p.String()))
	return uint32(uint32(h[7]) + uint32(h[4])<<8 + uint32(h[2])<<16 + uint32(h[3])<<24), nil
}

// AttrNames implements starlark.Value.
func (p *WifiNetworkProxy) AttrNames() []string {
	return []string{"mode", "set_mode", "ssid", "set_ssid", "psk", "set_psk", "disabled", "set_disabled"}
}

// Attr implements starlark.Value.
func (p *WifiNetworkProxy) Attr(name string) (starlark.Value, error) {
	switch name {
	case "mode":
		return starlark.MakeUint(uint(p.Conf.Mode)), nil
	case "set_mode":
		return starlark.NewBuiltin("set_mode", p.setMode), nil
	case "ssid":
		return starlark.String(p.Conf.SSID), nil
	case "set_ssid":
		return starlark.NewBuiltin("set_ssid", p.setSSID), nil
	case "psk":
		return starlark.String(p.Conf.PSK), nil
	case "set_psk":
		return starlark.NewBuiltin("set_psk", p.setPSK), nil
	case "disabled":
		return starlark.Bool(p.Conf.Disabled), nil
	case "set_disabled":
		return starlark.NewBuiltin("set_disabled", p.setDisabled), nil
	}

	return nil, starlark.NoSuchAttrError(
		fmt.Sprintf("%s has no .%s attribute", p.Type(), name))
}

func (p *WifiNetworkProxy) setSSID(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.SSID = string(s)
	return starlark.None, nil
}

func (p *WifiNetworkProxy) setPSK(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.PSK = string(s)
	return starlark.None, nil
}

func (p *WifiNetworkProxy) setDisabled(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	b, ok := args[0].(starlark.Bool)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.Disabled = bool(b)
	return starlark.None, nil
}

func (p *WifiNetworkProxy) setMode(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	i, ok := args[0].(starlark.Int)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	val, ok := i.Uint64()
	if !ok {
		return starlark.None, errors.New("argument must be an unsigned integer")
	}
	p.Conf.Mode = net.WPASupplicantMode(val)
	return starlark.None, nil
}

// SetField implements starlark.HasSetField.
func (p *WifiNetworkProxy) SetField(name string, val starlark.Value) error {
	switch name {
	case "mode":
		_, err := p.setMode(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "disabled":
		_, err := p.setDisabled(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "ssid":
		_, err := p.setSSID(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "psk":
		_, err := p.setPSK(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	}
	return errors.New("no such assignable field: " + name)
}

// WifiConfigurationProxy proxies a net.WPASupplicantConfig structure.
type WifiConfigurationProxy struct {
	Conf *net.WPASupplicantConfig
}

func (p *WifiConfigurationProxy) String() string {
	return p.Conf.String()
}

// Type implements starlark.Value.
func (p *WifiConfigurationProxy) Type() string {
	return fmt.Sprintf("net.WifiConfig")
}

// Freeze implements starlark.Value.
func (p *WifiConfigurationProxy) Freeze() {
}

// Truth implements starlark.Value.
func (p *WifiConfigurationProxy) Truth() starlark.Bool {
	return starlark.Bool(true)
}

// Hash implements starlark.Value.
func (p *WifiConfigurationProxy) Hash() (uint32, error) {
	h := sha256.Sum256([]byte(p.String()))
	return uint32(uint32(h[0]) + uint32(h[1])<<8 + uint32(h[2])<<16 + uint32(h[3])<<24), nil
}

// AttrNames implements starlark.Value.
func (p *WifiConfigurationProxy) AttrNames() []string {
	return []string{"control_interface", "set_control_interface", "control_interface_group", "set_control_interface_group",
		"allow_update_config", "set_allow_update_config", "country_code", "set_country_code",
		"device_name", "set_device_name", "networks", "set_networks"}
}

// Attr implements starlark.Value.
func (p *WifiConfigurationProxy) Attr(name string) (starlark.Value, error) {
	switch name {
	case "control_interface":
		return starlark.String(p.Conf.CtrlInterface), nil
	case "set_control_interface":
		return starlark.NewBuiltin("set_control_interface", p.setControlInterface), nil
	case "control_interface_group":
		return starlark.String(p.Conf.CtrlInterfaceGroup), nil
	case "set_control_interface_group":
		return starlark.NewBuiltin("set_control_interface_group", p.setControlInterfaceGroup), nil
	case "allow_update_config":
		return starlark.Bool(p.Conf.AllowUpdateConfig), nil
	case "set_allow_update_config":
		return starlark.NewBuiltin("set_allow_update_config", p.setAllowUpdateConfig), nil
	case "country_code":
		return starlark.String(p.Conf.CountryCode), nil
	case "set_country_code":
		return starlark.NewBuiltin("set_country_code", p.setCountryCode), nil
	case "device_name":
		return starlark.String(p.Conf.DeviceName), nil
	case "set_device_name":
		return starlark.NewBuiltin("set_device_name", p.setDeviceName), nil
	case "networks":
		var out []starlark.Value
		for _, n := range p.Conf.Networks {
			out = append(out, &WifiNetworkProxy{Conf: &n})
		}
		return starlark.NewList(out), nil
	case "set_networks":
		return starlark.NewBuiltin("set_networks", p.setNetworks), nil
	}

	return nil, starlark.NoSuchAttrError(
		fmt.Sprintf("%s has no .%s attribute", p.Type(), name))
}

func (p *WifiConfigurationProxy) setControlInterface(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.CtrlInterface = string(s)
	return starlark.None, nil
}

func (p *WifiConfigurationProxy) setControlInterfaceGroup(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.CtrlInterfaceGroup = string(s)
	return starlark.None, nil
}

func (p *WifiConfigurationProxy) setAllowUpdateConfig(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	b, ok := args[0].(starlark.Bool)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.AllowUpdateConfig = bool(b)
	return starlark.None, nil
}

func (p *WifiConfigurationProxy) setCountryCode(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.CountryCode = string(s)
	return starlark.None, nil
}

func (p *WifiConfigurationProxy) setDeviceName(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.DeviceName = string(s)
	return starlark.None, nil
}

func (p *WifiConfigurationProxy) setNetworks(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	networks, ok := args[0].(*starlark.List)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	out := make([]net.WPASupplicantNetwork, networks.Len())
	for i := 0; i < networks.Len(); i++ {
		n, ok := networks.Index(i).(*WifiNetworkProxy)
		if !ok {
			return starlark.None, fmt.Errorf("networks[%d] is not a WifiNetwork", i)
		}
		out[i] = *n.Conf
	}
	p.Conf.Networks = out
	return starlark.None, nil
}

// SetField implements starlark.HasSetField.
func (p *WifiConfigurationProxy) SetField(name string, val starlark.Value) error {
	switch name {
	case "control_interface":
		_, err := p.setControlInterface(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "control_interface_group":
		_, err := p.setControlInterfaceGroup(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "allow_update_config":
		_, err := p.setAllowUpdateConfig(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "country_code":
		_, err := p.setCountryCode(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "device_name":
		_, err := p.setDeviceName(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "networks":
		_, err := p.setNetworks(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	}
	return errors.New("no such assignable field: " + name)
}
