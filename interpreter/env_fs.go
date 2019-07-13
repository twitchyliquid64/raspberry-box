package interpreter

import (
	"os"

	"github.com/rekby/mbr"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	knownPartTypes = map[byte]string{
		0xc:  "FAT32-LBA",
		0x83: "Native Linux",
	}
)

func fsBuiltins(s *Script) starlark.StringDict {
	partEnums := starlark.StringDict{}
	for b, name := range knownPartTypes {
		partEnums[name] = starlark.MakeInt64(int64(b))
	}

	return starlark.StringDict{
		"exists": starlark.NewBuiltin("exists", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var path starlark.String
			if err := starlark.UnpackArgs("exists", args, kwargs, "path", &path); err != nil {
				return starlark.None, err
			}

			_, err := os.Stat(string(path))
			if err != nil && os.IsNotExist(err) {
				return starlark.Bool(false), nil
			}
			if err != nil {
				return starlark.None, err
			}
			return starlark.Bool(true), nil
		}),
		"enums": starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
			"partitions": starlarkstruct.FromStringDict(starlarkstruct.Default, partEnums),
		}),
		"read_partitions": makeReadPartitions(s),
		"stat": starlark.NewBuiltin("stat", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var path starlark.String
			if err := starlark.UnpackArgs("stat", args, kwargs, "path", &path); err != nil {
				return starlark.None, err
			}

			s, err := os.Stat(string(path))
			if err != nil {
				return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
					"success":    starlark.Bool(false),
					"error":      starlark.String(err.Error()),
					"not_exists": starlark.Bool(os.IsNotExist(err)),
				}), nil
			}
			return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
				"success":    starlark.Bool(true),
				"error":      starlark.Bool(false),
				"not_exists": starlark.Bool(false),
				"name":       starlark.String(s.Name()),
				"size":       starlark.MakeInt64(s.Size()),
				"dir":        starlark.Bool(s.IsDir()),
				"mode":       starlark.MakeInt64(int64(s.Mode())),
			}), nil
		}),
	}
}

func makeReadPartitions(s *Script) *starlark.Builtin {
	return starlark.NewBuiltin("read_partitions", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var path starlark.String
		if err := starlark.UnpackArgs("read_partitions", args, kwargs, "path", &path); err != nil {
			return starlark.None, err
		}

		f, err := os.Open(string(path))
		if err != nil {
			return starlark.None, err
		}
		tab, err := mbr.Read(f)
		if err != nil {
			return starlark.None, err
		}
		defer f.Close()

		var parts []starlark.Value
		for idx, p := range tab.GetAllPartitions() {
			parts = append(parts, starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
				"empty":     starlark.Bool(p.IsEmpty()),
				"bootable":  starlark.Bool(p.IsBootable()),
				"type_name": starlark.String(knownPartTypes[byte(p.GetType())]),
				"type":      starlark.MakeInt64(int64(p.GetType())),
				"index":     starlark.MakeInt64(int64(idx)),
				"lba": starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
					"length": starlark.MakeInt64(int64(p.GetLBALen())),
					"start":  starlark.MakeInt64(int64(p.GetLBAStart())),
				}),
			}))
		}

		return starlark.NewList(parts), nil
	})
}
