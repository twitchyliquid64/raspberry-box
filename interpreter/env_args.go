package interpreter

import (
	"errors"
	"flag"

	"go.starlark.net/starlark"
)

func argBuiltins(s *Script) starlark.StringDict {
	return starlark.StringDict{
		"verbose": starlark.Bool(s.verbose),
		"num_args": starlark.NewBuiltin("num_args", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			return starlark.MakeInt(s.fs.NArg()), nil
		}),
		"args": starlark.NewBuiltin("args", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var elements []starlark.Value
			for _, e := range s.fs.Args() {
				elements = append(elements, starlark.String(e))
			}
			return starlark.NewList(elements), nil
		}),
		"arg": starlark.NewBuiltin("arg", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var pos starlark.Int
			if err := starlark.UnpackArgs("arg", args, kwargs, "position", &pos); err != nil {
				return starlark.None, err
			}
			i, ok := pos.Int64()
			if !ok {
				return starlark.None, errors.New("cannot represent position as integer")
			}
			return starlark.String(s.fs.Arg(int(i))), nil
		}),
	}
}

func (s *Script) initFlags() error {
	s.fs = flag.NewFlagSet("raspberry-box", flag.ContinueOnError)

	if len(s.args) > 0 {
		if err := s.fs.Parse(s.args); err != nil {
			return err
		}
	}

	return nil
}
