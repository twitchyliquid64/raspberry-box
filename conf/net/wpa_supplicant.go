package net

import (
	"fmt"
	"strings"
)

// WPASupplicantConfig describes the configuration of the WPA supplicant
// daemon.
type WPASupplicantConfig struct {
	CtrlInterface      string // Path to specify control interface.
	CtrlInterfaceGroup string // Group allowed access to the control interface.
	AllowUpdateConfig  bool   // Generates update_config=1 if true.
	CountryCode        string // Such as US.
	DeviceName         string

	Networks []WPASupplicantNetwork
}

// String generates the config file.
func (c *WPASupplicantConfig) String() string {
	var out strings.Builder
	if c.CtrlInterface != "" {
		if c.CtrlInterfaceGroup != "" {
			out.WriteString("DIR=" + c.CtrlInterface + " GROUP=" + c.CtrlInterfaceGroup + "\n")
		}
		out.WriteString("ctrl_interface=" + c.CtrlInterface + "\n")
	}
	if c.AllowUpdateConfig {
		out.WriteString("update_config=1\n")
	} else {
		out.WriteString("update_config=0\n")
	}
	if c.CountryCode != "" {
		out.WriteString("country=" + c.CountryCode + "\n")
	}
	if c.DeviceName != "" {
		out.WriteString("device_name=" + c.DeviceName + "\n")
	}

	for i, n := range c.Networks {
		out.WriteString(n.String())

		if i < len(c.Networks)-1 {
			out.WriteString("\n")
		}
	}

	return out.String()
}

// WPASupplicantMode represents the type of network configuration.
type WPASupplicantMode uint8

// Valid WPASupplicantMode modes.
const (
	ModeClient WPASupplicantMode = iota
	ModeAdhoc
	ModeAP
)

// WPASupplicantNetwork represents configuration for a specific network.
type WPASupplicantNetwork struct {
	Mode     WPASupplicantMode
	Disabled bool
	SSID     string
	PSK      string
}

// String generates the config file section for the network.
func (c *WPASupplicantNetwork) String() string {
	var out strings.Builder

	out.WriteString("network={\n")
	out.WriteString(fmt.Sprintf("\tmode=%d\n", c.Mode))

	if c.Disabled {
		out.WriteString("\tdisabled=1\n")
	} else {
		out.WriteString("\tdisabled=0\n")
	}

	out.WriteString(fmt.Sprintf("\tssid=%q\n", c.SSID))
	out.WriteString(fmt.Sprintf("\tpsk=%q\n", c.PSK))
	out.WriteString("}\n")

	return out.String()
}
