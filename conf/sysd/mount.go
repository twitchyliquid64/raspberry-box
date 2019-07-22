package sysd

import (
	"fmt"
	"strings"
)

// Mount describes a mount unit.
type Mount struct {
	WhatPath     string   // Absolute path to device to mount.
	WherePath    string   // Absolute path to mount point.
	FSType       string   // Optional file-system path.
	MountOptions []string // Options to use when mounting.
}

// String returns the configuration as a valid mount stanza.
func (s *Mount) String() string {
	var out strings.Builder
	out.WriteString("[Mount]\n")

	if s.WhatPath != "" {
		out.WriteString(fmt.Sprintf("What=%s\n", s.WhatPath))
	}
	if s.WherePath != "" {
		out.WriteString(fmt.Sprintf("Where=%s\n", s.WherePath))
	}
	if s.FSType != "" {
		out.WriteString(fmt.Sprintf("Type=%s\n", s.FSType))
	}
	if len(s.MountOptions) > 0 {
		out.WriteString(fmt.Sprintf("Options=%s\n", strings.Join(s.MountOptions, ",")))
	}

	return out.String()
}
