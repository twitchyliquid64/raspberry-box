package interpreter

import (
	"errors"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func generateEnvironment(s *Script) (starlark.StringDict, error) {
	t, m := stdlibBuiltins()
	g := starlark.StringDict{
		"crash": starlark.NewBuiltin("crash", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			return nil, errors.New("soft crash")
		}),
		"struct":   starlark.NewBuiltin("struct", starlarkstruct.Make),
		"args":     starlarkstruct.FromStringDict(starlarkstruct.Default, argBuiltins(s)),
		"time":     starlarkstruct.FromStringDict(starlarkstruct.Default, t),
		"math":     starlarkstruct.FromStringDict(starlarkstruct.Default, m),
		"fs":       starlarkstruct.FromStringDict(starlarkstruct.Default, fsBuiltins(s)),
		"systemd":  starlarkstruct.FromStringDict(starlarkstruct.Default, sysdBuiltins(s)),
		"compiler": starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{"version": starlark.MakeInt64(starlark.CompilerVersion)}),
	}

	if s.testHook != nil {
		g["test_hook"] = starlark.NewBuiltin("test_hook", s.testHook)
	}

	return g, nil
}

// ScriptLoader provides a means for arbitrary imports to be resolved.
type ScriptLoader interface {
	resolveImport(name string) ([]byte, error)
}

func (s *Script) loadScript(script []byte, fname string, loader ScriptLoader) (*starlark.Thread, starlark.StringDict, error) {
	var moduleCache = map[string]starlark.StringDict{}
	var load func(_ *starlark.Thread, module string) (starlark.StringDict, error)
	predeclared, err := generateEnvironment(s)
	if err != nil {
		return nil, nil, err
	}

	load = func(_ *starlark.Thread, module string) (starlark.StringDict, error) {
		m, ok := moduleCache[module]
		if m == nil && ok {
			return nil, errors.New("cycle in dependency graph when loading " + module)
		}
		if m != nil {
			return m, nil
		}

		// loading in progress
		moduleCache[module] = nil
		d, err2 := loader.resolveImport(module)
		if err2 != nil {
			return nil, err2
		}
		thread := &starlark.Thread{
			Print: s.printFromSkylark,
			Load:  load,
		}
		mod, err2 := starlark.ExecFile(thread, module, d, predeclared)
		if err2 != nil {
			return nil, err2
		}
		moduleCache[module] = mod
		return mod, nil
	}

	thread := &starlark.Thread{
		Print: s.printFromSkylark,
		Load:  load,
	}

	globals, err := starlark.ExecFile(thread, fname, script, predeclared)
	if err != nil {
		return nil, nil, err
	}

	return thread, globals, nil
}
