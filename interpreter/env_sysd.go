package interpreter

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/twitchyliquid64/raspberry-box/conf/sysd"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func sysdBuiltins(s *Script) starlark.StringDict {
	return starlark.StringDict{
		"const": starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
			"restart_always":   starlark.String(sysd.RestartAlways),
			"restart_never":    starlark.String(sysd.RestartNever),
			"restart_success":  starlark.String(sysd.RestartOnSuccess),
			"restart_failure":  starlark.String(sysd.RestartOnFailure),
			"restart_abnormal": starlark.String(sysd.RestartOnAbnormal),
			"restart_watchdog": starlark.String(sysd.RestartOnWatchdog),
			"restart_abort":    starlark.String(sysd.RestartOnAbort),

			"killmode_controlgroup": starlark.String(sysd.KMControlGroup),

			"service_simple":  starlark.String(sysd.SimpleService),
			"service_exec":    starlark.String(sysd.ExecService),
			"service_forking": starlark.String(sysd.ForkingService),
			"service_oneshot": starlark.String(sysd.OneshotService),
			"service_notify":  starlark.String(sysd.NotifyService),
			"service_idle":    starlark.String(sysd.IdleService),

			"notifymode_none": starlark.String(sysd.NoNotify),
			"notifymode_main": starlark.String(sysd.NotifyMainProc),
			"notifymode_exec": starlark.String(sysd.NotifyExecProcs),
			"notifymode_all":  starlark.String(sysd.NotifyAllProcs),
		}),
		"out": starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
			"console": starlark.MakeInt64(int64(sysd.OutputConsole)),
			"journal": starlark.MakeInt64(int64(sysd.OutputJournal)),
			"inherit": starlark.MakeInt64(int64(sysd.OutputInherit)),
			"syslog":  starlark.MakeInt64(int64(sysd.OutputSyslog)),
			"kmesg":   starlark.MakeInt64(int64(sysd.OutputKmsg)),
		}),

		"Unit": starlark.NewBuiltin("Unit", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var description starlark.String
			var after, wantedBy, requiredBy *starlark.List
			var service starlark.Value
			if err := starlark.UnpackArgs("Unit", args, kwargs, "description?", &description, "after", &after, "wanted_by",
				&wantedBy, "required_by", &requiredBy, "service", &service); err != nil {
				return starlark.None, err
			}

			out := sysd.Unit{
				Description: string(description),
			}

			if service != nil {
				serv, ok := service.(*SystemdServiceProxy)
				if !ok {
					return starlark.None, fmt.Errorf("service parameter must be of type systemd.Service, got %T", service)
				}
				out.Service = serv.Service
				serv.Unit = &out
			}

			// Unpack after.
			if after != nil {
				for i := 0; i < after.Len(); i++ {
					s, ok := after.Index(i).(starlark.String)
					if !ok {
						return starlark.None, fmt.Errorf("after[%d] is not a string", i)
					}
					out.After = append(out.After, string(s))
				}
			}
			// Unpack wantedBy.
			if wantedBy != nil {
				for i := 0; i < wantedBy.Len(); i++ {
					s, ok := wantedBy.Index(i).(starlark.String)
					if !ok {
						return starlark.None, fmt.Errorf("wanted_by[%d] is not a string", i)
					}
					out.WantedBy = append(out.WantedBy, string(s))
				}
			}
			// Unpack requiredBy.
			if requiredBy != nil {
				for i := 0; i < requiredBy.Len(); i++ {
					s, ok := requiredBy.Index(i).(starlark.String)
					if !ok {
						return starlark.None, fmt.Errorf("required_by[%d] is not a string", i)
					}
					out.RequiredBy = append(out.RequiredBy, string(s))
				}
			}

			return &SystemdUnitProxy{
				Unit: &out,
			}, nil
		}),

		"Service": starlark.NewBuiltin("Service", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var (
				t, execStart, rootDir, usr, grp starlark.String
				killMode, restart               starlark.String
				restartSec, timeoutStopSec      starlark.Value
				watchdogSec                     starlark.Value
				ignoreSigpipe                   starlark.Bool
				stderr, stdout                  starlark.Int
			)
			if err := starlark.UnpackArgs("Service", args, kwargs, "type?", &t, "exec_start", &execStart,
				"root_dir", &rootDir, "user", &usr, "group", &grp, "restart", &restart,
				"kill_mode", &killMode, "timeout_stop_sec", &timeoutStopSec, "restart_sec", &restartSec,
				"watchdog_sec", &watchdogSec, "ignore_sigpipe", &ignoreSigpipe,
				"stderr", &stderr, "stdout", &stdout); err != nil {
				return starlark.None, err
			}

			out := sysd.Service{
				Type:          sysd.ServiceType(t),
				ExecStart:     string(execStart),
				RootDir:       string(rootDir),
				User:          string(usr),
				Group:         string(grp),
				KillMode:      sysd.KillMode(killMode),
				Restart:       sysd.RestartMode(restart),
				IgnoreSigpipe: bool(ignoreSigpipe),
			}
			if restartSec != nil {
				d, err := decodeDuration(restartSec)
				if err != nil {
					return starlark.None, fmt.Errorf("decoding restart_sec: %v", err)
				}
				out.RestartSec = d
			}
			if timeoutStopSec != nil {
				d, err := decodeDuration(timeoutStopSec)
				if err != nil {
					return starlark.None, fmt.Errorf("decoding timeout_stop_sec: %v", err)
				}
				out.TimeoutStopSec = d
			}
			if watchdogSec != nil {
				d, err := decodeDuration(watchdogSec)
				if err != nil {
					return starlark.None, fmt.Errorf("decoding watchdog_sec: %v", err)
				}
				out.WatchdogSec = d
			}
			if stderr, ok := stderr.Int64(); ok {
				out.Stderr = sysd.OutputSinks(stderr)
			}
			if stdout, ok := stderr.Int64(); ok {
				out.Stdout = sysd.OutputSinks(stdout)
			}

			return &SystemdServiceProxy{
				Service: &out,
			}, nil
		}),
	}
}

func decodeDuration(v starlark.Value) (time.Duration, error) {
	if num, ok := v.(starlark.Int); ok {
		uVal, _ := num.Uint64()
		return time.Duration(uVal), nil
	}
	if str, ok := v.(starlark.String); ok {
		return time.ParseDuration(string(str))
	}
	return time.Duration(0), fmt.Errorf("cannot represent type %T as a duration", v)
}

// SystemdUnitProxy proxies access to a unit structure.
type SystemdUnitProxy struct {
	Unit      *sysd.Unit
	servProxy *SystemdServiceProxy
}

func (p *SystemdUnitProxy) String() string {
	return p.String()
}

// Type implements starlark.Value.
func (p *SystemdUnitProxy) Type() string {
	return "systemd.Unit"
}

// Freeze implements starlark.Value.
func (p *SystemdUnitProxy) Freeze() {
}

// Truth implements starlark.Value.
func (p *SystemdUnitProxy) Truth() starlark.Bool {
	return starlark.Bool(p.Unit != nil)
}

// Hash implements starlark.Value.
func (p *SystemdUnitProxy) Hash() (uint32, error) {
	h := sha256.Sum256([]byte(p.String()))
	return uint32(uint32(h[0]) + uint32(h[1])<<8 + uint32(h[2])<<16 + uint32(h[3])<<24), nil
}

func (p *SystemdUnitProxy) setService(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(*SystemdServiceProxy)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has type %T", args[0])
	}
	p.servProxy = s
	p.Unit.Service = s.Service
	return starlark.None, nil
}

func (p *SystemdUnitProxy) setDescription(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Unit.Description = string(s)
	return starlark.None, nil
}

func (p *SystemdUnitProxy) appendAfter(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	for i, arg := range args {
		switch a := arg.(type) {
		case starlark.String:
			p.Unit.After = append(p.Unit.After, string(a))
		case *starlark.List:
			for x := 0; x < a.Len(); x++ {
				s, ok := a.Index(x).(starlark.String)
				if !ok {
					return starlark.None, fmt.Errorf("cannot handle argment %d index %d which has unhandled type %T", i, x, a.Index(x))
				}
				p.Unit.After = append(p.Unit.After, string(s))
			}
		default:
			return starlark.None, fmt.Errorf("cannot handle argment %d which has unhandled type %T", i, arg)
		}
	}

	return starlark.None, nil
}

func (p *SystemdUnitProxy) appendWantedBy(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	for i, arg := range args {
		switch a := arg.(type) {
		case starlark.String:
			p.Unit.WantedBy = append(p.Unit.WantedBy, string(a))
		case *starlark.List:
			for x := 0; x < a.Len(); x++ {
				s, ok := a.Index(x).(starlark.String)
				if !ok {
					return starlark.None, fmt.Errorf("cannot handle argment %d index %d which has unhandled type %T", i, x, a.Index(x))
				}
				p.Unit.WantedBy = append(p.Unit.WantedBy, string(s))
			}
		default:
			return starlark.None, fmt.Errorf("cannot handle argment %d which has unhandled type %T", i, arg)
		}
	}

	return starlark.None, nil
}

func (p *SystemdUnitProxy) appendRequiredBy(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	for i, arg := range args {
		switch a := arg.(type) {
		case starlark.String:
			p.Unit.RequiredBy = append(p.Unit.RequiredBy, string(a))
		case *starlark.List:
			for x := 0; x < a.Len(); x++ {
				s, ok := a.Index(x).(starlark.String)
				if !ok {
					return starlark.None, fmt.Errorf("cannot handle argment %d index %d which has unhandled type %T", i, x, a.Index(x))
				}
				p.Unit.RequiredBy = append(p.Unit.RequiredBy, string(s))
			}
		default:
			return starlark.None, fmt.Errorf("cannot handle argment %d which has unhandled type %T", i, arg)
		}
	}

	return starlark.None, nil
}

// Attr implements starlark.Value.
func (p *SystemdUnitProxy) Attr(name string) (starlark.Value, error) {
	switch name {
	case "description":
		return starlark.String(p.Unit.Description), nil
	case "set_description":
		return starlark.NewBuiltin("set_description", p.setDescription), nil

	case "after":
		return cvStrListToStarlark(p.Unit.After), nil
	case "append_after":
		return starlark.NewBuiltin("append_after", p.appendAfter), nil

	case "wanted_by":
		return cvStrListToStarlark(p.Unit.WantedBy), nil
	case "append_wanted_by":
		return starlark.NewBuiltin("append_wanted_by", p.appendWantedBy), nil

	case "required_by":
		return cvStrListToStarlark(p.Unit.RequiredBy), nil
	case "append_required_by":
		return starlark.NewBuiltin("append_required_by", p.appendRequiredBy), nil

	case "service":
		if p.Unit.Service == nil {
			return starlark.None, nil
		}
		if p.servProxy == nil {
			p.servProxy = &SystemdServiceProxy{
				Service: p.Unit.Service,
				Unit:    p.Unit,
			}
		}
		return p.servProxy, nil
	case "set_service":
		return starlark.NewBuiltin("set_service", p.setService), nil
	}

	return nil, starlark.NoSuchAttrError(
		fmt.Sprintf("%s has no .%s attribute", p.Type(), name))
}

// AttrNames implements starlark.Value.
func (p *SystemdUnitProxy) AttrNames() []string {
	return []string{"description", "set_description", "required_by", "append_required_by", "after", "append_after",
		"wanted_by", "append_wanted_by", "service", "set_service"}
}

// SetField implements starlark.HasSetField.
func (p *SystemdUnitProxy) SetField(name string, val starlark.Value) error {
	switch name {
	case "description":
		_, err := p.setDescription(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "service":
		s, ok := val.(*SystemdServiceProxy)
		if !ok {
			return fmt.Errorf("cannot assign value with type %T to a systemd.Service", val)
		}
		p.servProxy = s
		p.Unit.Service = s.Service
		return nil
	}
	return errors.New("no such assignable field: " + name)
}

// SystemdServiceProxy proxies access to a service structure.
type SystemdServiceProxy struct {
	Unit    *sysd.Unit
	Service *sysd.Service
}

func (p *SystemdServiceProxy) String() string {
	if p.Service == nil {
		return ""
	}
	return p.Service.String()
}

// Type implements starlark.Value.
func (p *SystemdServiceProxy) Type() string {
	return "systemd.Service"
}

// Freeze implements starlark.Value.
func (p *SystemdServiceProxy) Freeze() {
}

// Truth implements starlark.Value.
func (p *SystemdServiceProxy) Truth() starlark.Bool {
	return starlark.Bool(p.Service != nil)
}

// Hash implements starlark.Value.
func (p *SystemdServiceProxy) Hash() (uint32, error) {
	h := sha256.Sum256([]byte(p.String()))
	return uint32(uint32(h[0]) + uint32(h[4])<<8 + uint32(h[8])<<16 + uint32(h[9])<<24), nil
}

func (p *SystemdServiceProxy) setType(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.Type = sysd.ServiceType(s)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setExecStart(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.ExecStart = string(s)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setRootDir(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.RootDir = string(s)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setKillMode(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.KillMode = sysd.KillMode(s)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setUser(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.User = string(s)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setGroup(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.Group = string(s)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setRestart(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.Restart = sysd.RestartMode(s)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setRestartSec(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	d, err := decodeDuration(args[0])
	if err != nil {
		return starlark.None, err
	}
	p.Service.RestartSec = d
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setTimeoutStopSec(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	d, err := decodeDuration(args[0])
	if err != nil {
		return starlark.None, err
	}
	p.Service.TimeoutStopSec = d
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setWatchdogSec(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	d, err := decodeDuration(args[0])
	if err != nil {
		return starlark.None, err
	}
	p.Service.WatchdogSec = d
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setIgnoreSigpipe(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	b, ok := args[0].(starlark.Bool)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.IgnoreSigpipe = bool(b)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setStdout(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	b, ok := args[0].(starlark.Int)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	i, ok := b.Int64()
	if !ok {
		return starlark.None, errors.New("cannot represent argument as 64bit integer")
	}
	p.Service.Stdout = sysd.OutputSinks(i)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setStderr(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	b, ok := args[0].(starlark.Int)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	i, ok := b.Int64()
	if !ok {
		return starlark.None, errors.New("cannot represent argument as 64bit integer")
	}
	p.Service.Stderr = sysd.OutputSinks(i)
	return starlark.None, nil
}

// Attr implements starlark.Value.
func (p *SystemdServiceProxy) Attr(name string) (starlark.Value, error) {
	switch name {
	case "type":
		return starlark.String(p.Service.Type), nil
	case "set_type":
		return starlark.NewBuiltin("set_type", p.setType), nil
	case "exec_start":
		return starlark.String(p.Service.ExecStart), nil
	case "set_exec_start":
		return starlark.NewBuiltin("set_exec_start", p.setExecStart), nil
	case "root_dir":
		return starlark.String(p.Service.RootDir), nil
	case "set_root_dir":
		return starlark.NewBuiltin("set_root_dir", p.setRootDir), nil
	case "kill_mode":
		return starlark.String(p.Service.KillMode), nil
	case "set_kill_mode":
		return starlark.NewBuiltin("set_kill_mode", p.setKillMode), nil
	case "user":
		return starlark.String(p.Service.User), nil
	case "set_user":
		return starlark.NewBuiltin("set_user", p.setUser), nil
	case "group":
		return starlark.String(p.Service.Group), nil
	case "set_group":
		return starlark.NewBuiltin("set_user", p.setGroup), nil
	case "restart":
		return starlark.String(p.Service.Restart), nil
	case "set_restart":
		return starlark.NewBuiltin("set_restart", p.setRestart), nil
	case "restart_sec":
		return starlark.MakeUint64(uint64(p.Service.RestartSec)), nil
	case "set_restart_sec":
		return starlark.NewBuiltin("set_restart_sec", p.setRestartSec), nil
	case "timeout_stop_sec":
		return starlark.MakeUint64(uint64(p.Service.TimeoutStopSec)), nil
	case "set_timeout_stop_sec":
		return starlark.NewBuiltin("set_timeout_stop_sec", p.setTimeoutStopSec), nil
	case "watchdog_sec":
		return starlark.MakeUint64(uint64(p.Service.WatchdogSec)), nil
	case "set_watchdog_sec":
		return starlark.NewBuiltin("set_watchdog_sec", p.setWatchdogSec), nil
	case "ignore_sigpipe":
		return starlark.Bool(p.Service.IgnoreSigpipe), nil
	case "set_ignore_sigpipe":
		return starlark.NewBuiltin("set_ignore_sigpipe", p.setIgnoreSigpipe), nil
	case "stdout":
		return starlark.MakeInt64(int64(p.Service.Stdout)), nil
	case "set_stdout":
		return starlark.NewBuiltin("set_stdout", p.setStdout), nil
	case "stderr":
		return starlark.MakeInt64(int64(p.Service.Stdout)), nil
	case "set_stderr":
		return starlark.NewBuiltin("set_stderr", p.setStderr), nil
	}

	return nil, starlark.NoSuchAttrError(
		fmt.Sprintf("%s has no .%s attribute", p.Type(), name))
}

// SetField implements starlark.HasSetField.
func (p *SystemdServiceProxy) SetField(name string, val starlark.Value) error {
	switch name {
	case "user":
		_, err := p.setUser(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "group":
		_, err := p.setGroup(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "type":
		_, err := p.setType(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "kill_mode":
		_, err := p.setKillMode(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "exec_start":
		_, err := p.setExecStart(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "root_dir":
		_, err := p.setRootDir(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "restart":
		_, err := p.setRestart(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "restart_sec":
		_, err := p.setRestartSec(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "timeout_stop_sec":
		_, err := p.setTimeoutStopSec(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "watchdog_sec":
		_, err := p.setWatchdogSec(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "ignore_sigpipe":
		_, err := p.setIgnoreSigpipe(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "stdout":
		_, err := p.setStdout(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "stderr":
		_, err := p.setStderr(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	}
	return errors.New("no such assignable field: " + name)
}

// AttrNames implements starlark.Value.
func (p *SystemdServiceProxy) AttrNames() []string {
	return []string{"type", "set_type", "exec_start", "set_exec_start", "root_dir", "set_root_dir", "kill_mode", "set_kill_mode",
		"user", "set_user", "group", "set_group", "restart", "set_restart", "restart_sec", "set_timeout_stop_sec", "timeout_stop_sec",
		"set_watchdog_sec", "watchdog_sec", "set_ignore_sigpipe", "ignore_sigpipe", "stdout", "set_stdout", "stderr", "set_stderr"}
}
