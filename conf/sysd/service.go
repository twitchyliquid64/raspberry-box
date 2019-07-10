package sysd

import (
	"fmt"
	"strings"
	"time"
)

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
	RestartAlways     RestartMode = "always"
	RestartNever      RestartMode = "no"
	RestartOnSuccess  RestartMode = "on-success"
	RestartOnFailure  RestartMode = "on-failure"
	RestartOnAbnormal RestartMode = "on-abnormal"
	RestartOnWatchdog RestartMode = "on-watchdog"
	RestartOnAbort    RestartMode = "on-abort"
)

// ServiceType describes the type of systemd service.
type ServiceType string

const (
	// SimpleService is a service that launches a process. The service is
	// considered started once the unit initalizes, regardless of if the process
	// started successfully or not.
	SimpleService ServiceType = "simple"
	// ExecService is a service that launches a process, but blocks till the
	// fork and exec have happened successfully. The unit fails if either of
	// those system calls failed.
	ExecService ServiceType = "exec"
	// ForkingService is expected to fork itself then exit. Systemd will consider
	// the service started when the process the unit started exits.
	ForkingService ServiceType = "forking"
	// OneshotService represents a service which is considered started once the
	// main process exits.
	OneshotService ServiceType = "oneshot"
	// NotifyService is a service that launches a process. The process must notify
	// systemd when it has initialized.
	NotifyService ServiceType = "notify"
	// IdleService is a service that launches a process, but the launch is delayed
	// until active jobs are dispatched or 5s has elapsed.
	IdleService ServiceType = "idle"
)

// NotifySockMode describes access to the service status notification socket.
type NotifySockMode string

// Valid NotifySockMode values.
const (
	NoNotify        NotifySockMode = "none"
	NotifyMainProc  NotifySockMode = "main"
	NotifyExecProcs NotifySockMode = "exec"
	NotifyAllProcs  NotifySockMode = "all"
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
			mask:  OutputSyslog,
			ident: "syslog",
		},
		{
			mask:  OutputKmsg,
			ident: "kmsg",
		},
		{
			mask:  OutputJournal,
			ident: "journal",
		},
		{
			mask:  OutputConsole,
			ident: "console",
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
	OutputSyslog
	OutputKmsg
)

// Service represents the configuration of a systemd service.
type Service struct {
	Type        ServiceType
	ExecStart   string
	RootDir     string
	KillMode    KillMode
	User, Group string

	TimeoutStopSec time.Duration
	Restart        RestartMode
	RestartSec     time.Duration
	WatchdogSec    time.Duration

	NotifyAccess NotifySockMode

	IgnoreSigpipe bool
	Stdout        OutputSinks
	Stderr        OutputSinks

	Conditions Conditions
}

// String returns the configuration as a valid service stanza.
func (s *Service) String() string {
	var out strings.Builder
	out.WriteString("[Service]\n")

	if s.Type != "" {
		out.WriteString(fmt.Sprintf("Type=%s\n", s.Type))
	}
	if s.ExecStart != "" {
		out.WriteString(fmt.Sprintf("ExecStart=%s\n", s.ExecStart))
	}
	if s.RootDir != "" {
		out.WriteString(fmt.Sprintf("RootDirectory=%s\n", s.RootDir))
	}
	if s.KillMode != "" {
		out.WriteString(fmt.Sprintf("KillMode=%s\n", s.KillMode))
	}

	if s.User != "" {
		out.WriteString(fmt.Sprintf("User=%s\n", s.User))
	}
	if s.Group != "" {
		out.WriteString(fmt.Sprintf("Group=%s\n", s.Group))
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
	if s.WatchdogSec > 0 {
		out.WriteString(fmt.Sprintf("WatchdogSec=%s\n", s.WatchdogSec.String()))
	}

	if s.NotifyAccess != "" {
		out.WriteString(fmt.Sprintf("NotifyAccess=%s\n", s.NotifyAccess))
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

	if cond := s.Conditions.String(); len(cond) > 0 {
		out.WriteString(cond)
	}

	return out.String()
}
