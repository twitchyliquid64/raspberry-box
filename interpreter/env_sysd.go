package interpreter

import (
	"fmt"
	"time"

	"github.com/twitchyliquid64/raspberry-box/conf/sysd"
	sd "github.com/twitchyliquid64/raspberry-box/sysd"
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

		"Mount": starlark.NewBuiltin("Mount", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var whatPath, wherePath, fsType starlark.String
			if err := starlark.UnpackArgs("Mount", args, kwargs, "what_path?", &whatPath, "where_path", &wherePath, "fs_type",
				&fsType); err != nil {
				return starlark.None, err
			}
			out := sysd.Mount{
				WhatPath:  string(whatPath),
				WherePath: string(wherePath),
				FSType:    string(fsType),
			}
			return &SystemdMountProxy{
				Conf: &out,
			}, nil
		}),

		"Service": starlark.NewBuiltin("Service", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var (
				t, execStart, rootDir, usr, grp  starlark.String
				execReload, execStop, workingDir starlark.String
				execStartPre, execStopPost       starlark.String
				killMode, restart                starlark.String
				restartSec, timeoutStopSec       starlark.Value
				watchdogSec                      starlark.Value
				ignoreSigpipe                    starlark.Bool
				stderr, stdout                   starlark.Int
			)
			if err := starlark.UnpackArgs("Service", args, kwargs, "type?", &t, "exec_start", &execStart, "working_dir", &workingDir,
				"root_dir", &rootDir, "user", &usr, "group", &grp, "exec_reload", &execReload, "exec_stop", &execStop,
				"exec_start_pre", &execStartPre, "exec_stop_post", &execStopPost, "restart", &restart,
				"kill_mode", &killMode, "timeout_stop_sec", &timeoutStopSec, "restart_sec", &restartSec,
				"watchdog_sec", &watchdogSec, "ignore_sigpipe", &ignoreSigpipe,
				"stderr", &stderr, "stdout", &stdout); err != nil {
				return starlark.None, err
			}

			out := sysd.Service{
				Type:          sysd.ServiceType(t),
				ExecStart:     string(execStart),
				RootDir:       string(rootDir),
				WorkingDir:    string(workingDir),
				User:          string(usr),
				Group:         string(grp),
				KillMode:      sysd.KillMode(killMode),
				ExecReload:    string(execReload),
				ExecStop:      string(execStop),
				ExecStartPre:  string(execStartPre),
				ExecStopPost:  string(execStopPost),
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
		"ConditionExists": starlark.NewBuiltin("ConditionExists", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var path starlark.String
			if err := starlark.UnpackPositionalArgs("ConditionExists", args, kwargs, 1, &path); err != nil {
				return starlark.None, err
			}
			return &SystemdConditionProxy{
				Kind: "ConditionPathExists",
				Arg:  string(path),
			}, nil
		}),
		"ConditionNotExists": starlark.NewBuiltin("ConditionNotExists", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var path starlark.String
			if err := starlark.UnpackPositionalArgs("ConditionNotExists", args, kwargs, 1, &path); err != nil {
				return starlark.None, err
			}
			return &SystemdConditionProxy{
				Kind: "ConditionPathNotExists",
				Arg:  string(path),
			}, nil
		}),
		"ConditionHost": starlark.NewBuiltin("ConditionHost", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var path starlark.String
			if err := starlark.UnpackPositionalArgs("ConditionHost", args, kwargs, 1, &path); err != nil {
				return starlark.None, err
			}
			return &SystemdConditionProxy{
				Kind: "ConditionHost",
				Arg:  string(path),
			}, nil
		}),
		"ConditionFirstBoot": starlark.NewBuiltin("ConditionFirstBoot", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var path starlark.String
			if err := starlark.UnpackPositionalArgs("ConditionFirstBoot", args, kwargs, 1, &path); err != nil {
				return starlark.None, err
			}
			return &SystemdConditionProxy{
				Kind: "ConditionFirstBoot",
				Arg:  string(path),
			}, nil
		}),
		"install": starlark.NewBuiltin("install", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var name starlark.String
			var f, u starlark.Value
			if err := starlark.UnpackArgs("install", args, kwargs, "fs", &f, "name", &name, "unit", &u); err != nil {
				return starlark.None, err
			}

			fs, ok := f.(*FSMountProxy)
			if !ok {
				return starlark.None, fmt.Errorf("unit parameter must be of type systemd.Unit, got %T", f)
			}

			unit, ok := u.(*SystemdUnitProxy)
			if !ok {
				return starlark.None, fmt.Errorf("unit parameter must be of type systemd.Unit, got %T", u)
			}

			return starlark.None, sd.Install(fs.fs, string(name), unit.Unit, true)
		}),
		"is_installed": starlark.NewBuiltin("is_installed", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var unit starlark.String
			var f starlark.Value
			if err := starlark.UnpackArgs("is_installed", args, kwargs, "fs", &f, "unit", &unit); err != nil {
				return starlark.None, err
			}

			fs, ok := f.(*FSMountProxy)
			if !ok {
				return starlark.None, fmt.Errorf("fs parameter must be of type fs.Mount, got %T", f)
			}
			enabled, err := sd.Exists(fs.fs, string(unit))
			if err != nil {
				return starlark.None, err
			}
			return starlark.Bool(enabled), nil
		}),
		"enable_target": starlark.NewBuiltin("enable_target", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var unit, target starlark.String
			var f starlark.Value
			if err := starlark.UnpackArgs("enable_target", args, kwargs, "fs", &f, "unit", &unit, "target", &target); err != nil {
				return starlark.None, err
			}

			fs, ok := f.(*FSMountProxy)
			if !ok {
				return starlark.None, fmt.Errorf("fs parameter must be of type fs.Mount, got %T", f)
			}
			return starlark.None, sd.Enable(fs.fs, string(unit), string(target))
		}),
		"is_enabled_on_target": starlark.NewBuiltin("is_enabled_on_target", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var unit, target starlark.String
			var f starlark.Value
			if err := starlark.UnpackArgs("is_enabled_on_target", args, kwargs, "fs", &f, "unit", &unit, "target", &target); err != nil {
				return starlark.None, err
			}

			fs, ok := f.(*FSMountProxy)
			if !ok {
				return starlark.None, fmt.Errorf("fs parameter must be of type fs.Mount, got %T", f)
			}
			enabled, err := sd.IsEnabledOnTarget(fs.fs, string(unit), string(target))
			if err != nil {
				return starlark.None, err
			}
			return starlark.Bool(enabled), nil
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
