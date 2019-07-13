package interpreter

import (
	"errors"
	"time"

	"go.starlark.net/starlark"
)

func stdlibBuiltins() (starlark.StringDict, starlark.StringDict) {
	timeStruct := starlark.StringDict{
		"start": starlark.MakeInt64(time.Now().UnixNano()),
		"now": starlark.NewBuiltin("now", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			return starlark.MakeInt64(time.Now().UnixNano()), nil
		}),
	}

	mathStruct := starlark.StringDict{
		"shl": starlark.NewBuiltin("shl", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var base starlark.Int
			var shiftAmount int
			if err := starlark.UnpackArgs("shl", args, kwargs, "base", &base, "shift", &shiftAmount); err != nil {
				return starlark.None, err
			}
			b, ok := base.Uint64()
			if !ok {
				return starlark.None, errors.New("cannot represent base as unsigned integer")
			}
			return starlark.MakeUint64(b << uint(shiftAmount)), nil
		}),
		"shr": starlark.NewBuiltin("shr", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var base starlark.Int
			var shiftAmount int
			if err := starlark.UnpackArgs("shr", args, kwargs, "base", &base, "shift", &shiftAmount); err != nil {
				return starlark.None, err
			}
			b, ok := base.Uint64()
			if !ok {
				return starlark.None, errors.New("cannot represent base as unsigned integer")
			}
			return starlark.MakeUint64(b >> uint(shiftAmount)), nil
		}),
		"_not": starlark.NewBuiltin("_not", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var base starlark.Int
			if err := starlark.UnpackArgs("_not", args, kwargs, "base", &base); err != nil {
				return starlark.None, err
			}
			b, ok := base.Uint64()
			if !ok {
				return starlark.None, errors.New("cannot represent base as unsigned integer")
			}
			return starlark.MakeUint64(^b), nil
		}),
		"_and": starlark.NewBuiltin("_and", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var o1, o2 starlark.Int
			if err := starlark.UnpackArgs("_and", args, kwargs, "op1", &o1, "op2", &o2); err != nil {
				return starlark.None, err
			}
			i1, ok := o1.Uint64()
			if !ok {
				return starlark.None, errors.New("cannot represent op1 as unsigned integer")
			}
			i2, ok := o2.Uint64()
			if !ok {
				return starlark.None, errors.New("cannot represent op2 as unsigned integer")
			}
			return starlark.MakeUint64(i1 & i2), nil
		}),
	}

	return timeStruct, mathStruct
}
