package interpreter

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/twitchyliquid64/raspberry-box/conf/sysd"
	"go.starlark.net/starlark"
)

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

func (p *SystemdServiceProxy) setWorkingDir(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.WorkingDir = string(s)
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

func (p *SystemdServiceProxy) setExecReload(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.ExecReload = string(s)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setExecStop(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.ExecStop = string(s)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setExecStartPre(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.ExecStartPre = string(s)
	return starlark.None, nil
}

func (p *SystemdServiceProxy) setExecStopPost(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Service.ExecStopPost = string(s)
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

func (p *SystemdServiceProxy) setConditions(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	l, ok := args[0].(*starlark.List)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}

	var conditions sysd.Conditions
	for x := 0; x < l.Len(); x++ {
		c, ok := l.Index(x).(*SystemdConditionProxy)
		if !ok {
			return starlark.None, fmt.Errorf("cannot handle index %d which has unhandled type %T", x, l.Index(x))
		}
		switch c.Kind {
		case "ConditionPathExists":
			conditions = append(conditions, sysd.ConditionExists(c.Arg))
		case "ConditionPathNotExists":
			conditions = append(conditions, sysd.ConditionNotExists(c.Arg))
		case "ConditionHost":
			conditions = append(conditions, sysd.ConditionHost(c.Arg))
		default:
			return starlark.None, fmt.Errorf("index %d has unknown condition kind %s", x, c.Kind)
		}
	}
	p.Service.Conditions = conditions

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
	case "exec_reload":
		return starlark.String(p.Service.ExecReload), nil
	case "set_exec_reload":
		return starlark.NewBuiltin("set_exec_reload", p.setExecReload), nil
	case "exec_stop":
		return starlark.String(p.Service.ExecStop), nil
	case "set_exec_stop":
		return starlark.NewBuiltin("set_exec_stop", p.setExecStop), nil
	case "exec_start_pre":
		return starlark.String(p.Service.ExecStartPre), nil
	case "set_exec_start_pre":
		return starlark.NewBuiltin("set_exec_start_pre", p.setExecStartPre), nil
	case "exec_stop_post":
		return starlark.String(p.Service.ExecStopPost), nil
	case "set_exec_stop_post":
		return starlark.NewBuiltin("set_exec_stop_post", p.setExecStopPost), nil

	case "working_dir":
		return starlark.String(p.Service.WorkingDir), nil
	case "set_working_dir":
		return starlark.NewBuiltin("set_working_dir", p.setWorkingDir), nil
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
	case "conditions":
		var out []starlark.Value
		for _, c := range p.Service.Conditions {
			switch cond := c.(type) {
			case sysd.ConditionExists:
				out = append(out, &SystemdConditionProxy{Kind: "ConditionPathExists", Arg: string(cond)})
			case sysd.ConditionNotExists:
				out = append(out, &SystemdConditionProxy{Kind: "ConditionPathNotExists", Arg: string(cond)})
			case sysd.ConditionHost:
				out = append(out, &SystemdConditionProxy{Kind: "ConditionHost", Arg: string(cond)})
			default:
				return starlark.None, fmt.Errorf("unknown condition %T", c)
			}
		}
		return starlark.NewList(out), nil
	case "set_conditions":
		return starlark.NewBuiltin("set_conditions", p.setConditions), nil
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
	case "exec_reload":
		_, err := p.setExecReload(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "exec_stop":
		_, err := p.setExecStop(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "exec_start_pre":
		_, err := p.setExecStartPre(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "exec_stop_post":
		_, err := p.setExecStopPost(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "working_dir":
		_, err := p.setWorkingDir(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
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
	case "conditions":
		_, err := p.setConditions(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	}
	return errors.New("no such assignable field: " + name)
}

// AttrNames implements starlark.Value.
func (p *SystemdServiceProxy) AttrNames() []string {
	return []string{"type", "set_type", "exec_start", "set_exec_start", "root_dir", "set_root_dir", "kill_mode", "set_kill_mode", "working_dir",
		"user", "set_user", "group", "set_group", "exec_reload", "set_exec_reload", "exec_stop", "set_exec_stop", "exec_start_pre", "set_exec_start_pre",
		"exec_stop_post", "set_exec_stop_post", "restart", "set_restart", "restart_sec", "set_timeout_stop_sec", "timeout_stop_sec",
		"set_watchdog_sec", "watchdog_sec", "set_ignore_sigpipe", "ignore_sigpipe", "stdout", "set_stdout", "stderr", "set_stderr",
		"conditions", "set_conditions"}
}

// SystemdConditionProxy proxies access to a condition structure.
type SystemdConditionProxy struct {
	Kind string
	Arg  string
}

func (p *SystemdConditionProxy) String() string {
	return fmt.Sprintf("systemd.%s{%s}", p.Kind, p.Arg)
}

// Type implements starlark.Value.
func (p *SystemdConditionProxy) Type() string {
	return "systemd.Condition" + p.Kind
}

// Freeze implements starlark.Value.
func (p *SystemdConditionProxy) Freeze() {
}

// Truth implements starlark.Value.
func (p *SystemdConditionProxy) Truth() starlark.Bool {
	return starlark.Bool(true)
}

// Hash implements starlark.Value.
func (p *SystemdConditionProxy) Hash() (uint32, error) {
	h := sha256.Sum256([]byte(p.String()))
	return uint32(uint32(h[0]) + uint32(h[1])<<8 + uint32(h[2])<<16 + uint32(h[3])<<24), nil
}

// AttrNames implements starlark.Value.
func (p *SystemdConditionProxy) AttrNames() []string {
	return []string{"arg"}
}

// Attr implements starlark.Value.
func (p *SystemdConditionProxy) Attr(name string) (starlark.Value, error) {
	switch name {
	case "arg":
		return starlark.String(p.Arg), nil
	}

	return nil, starlark.NoSuchAttrError(
		fmt.Sprintf("%s has no .%s attribute", p.Type(), name))
}

// SystemdMountProxy proxies access to a mount structure.
type SystemdMountProxy struct {
	Conf *sysd.Mount
}

func (p *SystemdMountProxy) String() string {
	return fmt.Sprintf("systemd.Mount{%p}", p)
}

// Type implements starlark.Value.
func (p *SystemdMountProxy) Type() string {
	return "systemd.Mount"
}

// Freeze implements starlark.Value.
func (p *SystemdMountProxy) Freeze() {
}

// Truth implements starlark.Value.
func (p *SystemdMountProxy) Truth() starlark.Bool {
	return starlark.Bool(true)
}

// Hash implements starlark.Value.
func (p *SystemdMountProxy) Hash() (uint32, error) {
	h := sha256.Sum256([]byte(p.String()))
	return uint32(uint32(h[0]) + uint32(h[1])<<8 + uint32(h[2])<<16 + uint32(h[3])<<24), nil
}

// AttrNames implements starlark.Value.
func (p *SystemdMountProxy) AttrNames() []string {
	return []string{"what_path", "where_path", "fs_type", "set_what_path", "set_where_path", "set_fs_type"}
}

func (p *SystemdMountProxy) setWhatPath(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.WhatPath = string(s)
	return starlark.None, nil
}

func (p *SystemdMountProxy) setWherePath(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.WherePath = string(s)
	return starlark.None, nil
}

func (p *SystemdMountProxy) setFSType(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, ok := args[0].(starlark.String)
	if !ok {
		return starlark.None, fmt.Errorf("cannot handle argument 0 which has unhandled type %T", args[0])
	}
	p.Conf.FSType = string(s)
	return starlark.None, nil
}

// Attr implements starlark.Value.
func (p *SystemdMountProxy) Attr(name string) (starlark.Value, error) {
	switch name {
	case "what_path":
		return starlark.String(p.Conf.WhatPath), nil
	case "set_what_path":
		return starlark.NewBuiltin("set_what_path", p.setWhatPath), nil
	case "where_path":
		return starlark.String(p.Conf.WherePath), nil
	case "set_where_path":
		return starlark.NewBuiltin("set_where_path", p.setWherePath), nil
	case "fs_type":
		return starlark.String(p.Conf.FSType), nil
	case "set_fs_type":
		return starlark.NewBuiltin("set_fs_type", p.setFSType), nil
	}

	return nil, starlark.NoSuchAttrError(
		fmt.Sprintf("%s has no .%s attribute", p.Type(), name))
}

// SetField implements starlark.HasSetField.
func (p *SystemdMountProxy) SetField(name string, val starlark.Value) error {
	switch name {
	case "what_path":
		_, err := p.setWhatPath(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "where_path":
		_, err := p.setWherePath(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	case "fs_type":
		_, err := p.setFSType(nil, nil, starlark.Tuple([]starlark.Value{val}), nil)
		return err
	}
	return errors.New("no such assignable field: " + name)
}
