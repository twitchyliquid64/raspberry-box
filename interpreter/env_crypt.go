package interpreter

import (
	"github.com/tredoe/osutil/user/crypt"
	_ "github.com/tredoe/osutil/user/crypt/sha512_crypt"
	"go.starlark.net/starlark"
)

func cryptBuiltins(s *Script) starlark.StringDict {
	return starlark.StringDict{
		"unix_hash": starlark.NewBuiltin("unix_hash", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var pw starlark.String
			if err := starlark.UnpackArgs("unix_hash", args, kwargs, "password", &pw); err != nil {
				return starlark.None, err
			}

			s, err := crypt.New(crypt.SHA512).Generate([]byte(pw), nil)
			if err != nil {
				return starlark.None, err
			}
			return starlark.String(s), nil
		}),
	}
}
