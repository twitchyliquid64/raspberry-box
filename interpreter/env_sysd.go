package interpreter

import (
	"crypto/sha256"
	"fmt"

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
			var t, execStart, rootDir, usr, grp, restart starlark.String
			if err := starlark.UnpackArgs("Service", args, kwargs, "type?", &t, "exec_start", &execStart,
				"root_dir", &rootDir, "user", &usr, "group", &grp, "restart", &restart); err != nil {
				return starlark.None, err
			}

			out := sysd.Service{
				Type:      sysd.ServiceType(t),
				ExecStart: string(execStart),
				RootDir:   string(rootDir),
				User:      string(usr),
				Group:     string(grp),
				Restart:   sysd.RestartMode(restart),
			}

			return &SystemdServiceProxy{
				Service: &out,
			}, nil
		}),
	}
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
	}

	return nil, starlark.NoSuchAttrError(
		fmt.Sprintf("%s has no .%s attribute", p.Type(), name))
}

// AttrNames implements starlark.Value.
func (p *SystemdServiceProxy) AttrNames() []string {
	return []string{"type", "set_type", "exec_start", "set_exec_start", "root_dir", "set_root_dir", "kill_mode", "set_kill_mode",
		"user", "set_user", "group", "set_group"}
}
