// Package sysd generates systemd config files.
package sysd

import (
	"fmt"
	"strings"
)

// Unit represents the configuration of a systemd unit.
type Unit struct {
	Description string
	After       []string

	Service *Service

	WantedBy   []string
	RequiredBy []string
}

// String returns the structure represent in the correct file format.
func (u *Unit) String() string {
	var out strings.Builder
	out.WriteString("[Unit]\n")
	if u.Description != "" {
		out.WriteString(fmt.Sprintf("Description=%s\n", u.Description))
	}
	if len(u.After) > 0 {
		out.WriteString(fmt.Sprintf("After=%s\n", strings.Join(u.After, " ")))
	}
	out.WriteString("\n")

	if u.Service != nil {
		out.WriteString(u.Service.String())
		out.WriteString("\n")
	}

	if len(u.WantedBy) > 0 || len(u.RequiredBy) > 0 {
		out.WriteString("[Install]\n")
		if len(u.WantedBy) > 0 {
			out.WriteString(fmt.Sprintf("WantedBy=%s\n", strings.Join(u.WantedBy, " ")))
		}
		if len(u.RequiredBy) > 0 {
			out.WriteString(fmt.Sprintf("RequiredBy=%s\n", strings.Join(u.RequiredBy, " ")))
		}
		out.WriteString("\n")
	}

	return out.String()
}
