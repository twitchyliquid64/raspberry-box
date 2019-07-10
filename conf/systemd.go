// Package conf generates configuration files.
package conf

import (
	"fmt"
	"strings"
	"time"
)

// SystemdUnit represents the configuration of a systemd unit.
type SystemdUnit struct {
	Description string
	After       []string

	Service *SystemdService

	WantedBy   []string
	RequiredBy []string
}

// String returns the structure represent in the correct file format.
func (u *SystemdUnit) String() string {
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

// KillMode describes kill modes of a systemd service.
type KillMode string

// Valid kill modes.
const (
	KMControlGroup KillMode = "control-group"
)

// RestartMode describes how/if a systemd service should be restarted.
type RestartMode string

// Valid restart modes.
const (
	RestartAlways RestartMode = "always"
)

// OutputSinks describes where standard output should be written.
type OutputSinks uint8

func (o OutputSinks) String() string {
	var (
		out strings.Builder
		i   int
	)

	for _, opt := range []struct {
		mask  OutputSinks
		ident string
	}{
		{
			mask:  OutputConsole,
			ident: "console",
		},
		{
			mask:  OutputJournal,
			ident: "journal",
		},
		{
			mask:  OutputInherit,
			ident: "inherit",
		},
	} {
		if opt.mask&OutputConsole != 0 {
			if i > 0 {
				out.WriteString("+")
			}
			out.WriteString(opt.ident)
			i++
		}
	}

	return out.String()
}

// Output sink flags.
const (
	OutputConsole OutputSinks = 1 << iota
	OutputJournal
	OutputInherit
)

// SystemdService represents the configuration of a systemd service.
type SystemdService struct {
	ExecStart string
	KillMode  KillMode

	TimeoutStopSec time.Duration
	Restart        RestartMode
	RestartSec     time.Duration

	IgnoreSigpipe bool
	Stdout        OutputSinks
	Stderr        OutputSinks
}

// String returns the configuration as a valid service stanza.
func (s *SystemdService) String() string {
	var out strings.Builder
	out.WriteString("[Service]\n")

	if s.ExecStart != "" {
		out.WriteString(fmt.Sprintf("ExecStart=%s\n", s.ExecStart))
	}
	if s.KillMode != "" {
		out.WriteString(fmt.Sprintf("KillMode=%s\n", s.KillMode))
	}

	if s.TimeoutStopSec > 0 {
		out.WriteString(fmt.Sprintf("TimeoutStopSec=%s\n", s.TimeoutStopSec.String()))
	}
	if s.Restart != "" {
		out.WriteString(fmt.Sprintf("Restart=%s\n", s.Restart))
	}
	if s.RestartSec > 0 {
		out.WriteString(fmt.Sprintf("RestartSec=%s\n", s.RestartSec.String()))
	}

	if s.IgnoreSigpipe {
		out.WriteString("IgnoreSIGPIPE=yes\n")
	} else {
		out.WriteString("IgnoreSIGPIPE=no\n")
	}

	if s.Stdout != 0 {
		out.WriteString(fmt.Sprintf("StandardOutput=%s\n", s.Stdout.String()))
	}
	if s.Stderr != 0 {
		out.WriteString(fmt.Sprintf("StandardError=%s\n", s.Stderr.String()))
	}

	return out.String()
}
