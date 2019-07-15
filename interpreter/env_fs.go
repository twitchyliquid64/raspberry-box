package interpreter

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"

	"github.com/rekby/mbr"
	"github.com/twitchyliquid64/raspberry-box/fs"
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
		"mnt_ext4": starlark.NewBuiltin("mnt_ext4", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var path starlark.String
			var part *starlarkstruct.Struct
			if err := starlark.UnpackPositionalArgs("mnt_ext4", args, kwargs, 2, &path, &part); err != nil {
				return starlark.None, err
			}
			if part == nil {
				return starlark.None, errors.New("no partition information provided")
			}

			// Unpack partition information.
			tmp, err := part.Attr("lba")
			if err != nil {
				return starlark.None, err
			}
			lba := tmp.(*starlarkstruct.Struct)
			start, err := lba.Attr("start")
			if err != nil {
				return starlark.None, err
			}
			length, err := lba.Attr("length")
			if err != nil {
				return starlark.None, err
			}

			st, ok := start.(starlark.Int).Int64()
			if !ok {
				return starlark.None, errors.New("start is not an integer")
			}
			l, ok := length.(starlark.Int).Int64()
			if !ok {
				return starlark.None, errors.New("length is not an integer")
			}

			mnt, err := fs.KMountExt4(string(path), uint64(st)*512, uint64(l)*512)
			if err != nil {
				return starlark.None, err
			}
			out := &FSMountProxy{
				Kind: "Ext4",
				Path: string(path),
				fs:   mnt,
			}

			s.resources = append(s.resources, out)
			return out, nil
		}),
		"mnt_vfat": starlark.NewBuiltin("mnt_vfat", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var path starlark.String
			var part *starlarkstruct.Struct
			if err := starlark.UnpackPositionalArgs("mnt_vfat", args, kwargs, 2, &path, &part); err != nil {
				return starlark.None, err
			}
			if part == nil {
				return starlark.None, errors.New("no partition information provided")
			}

			// Unpack partition information.
			tmp, err := part.Attr("lba")
			if err != nil {
				return starlark.None, err
			}
			lba := tmp.(*starlarkstruct.Struct)
			start, err := lba.Attr("start")
			if err != nil {
				return starlark.None, err
			}
			length, err := lba.Attr("length")
			if err != nil {
				return starlark.None, err
			}

			st, ok := start.(starlark.Int).Int64()
			if !ok {
				return starlark.None, errors.New("start is not an integer")
			}
			l, ok := length.(starlark.Int).Int64()
			if !ok {
				return starlark.None, errors.New("length is not an integer")
			}

			mnt, err := fs.KMountVFat(string(path), uint64(st)*512, uint64(l)*512)
			if err != nil {
				return starlark.None, err
			}
			out := &FSMountProxy{
				Kind: "VFAT",
				Path: string(path),
				fs:   mnt,
			}

			s.resources = append(s.resources, out)
			return out, nil
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

// FS describes an interface to the filesystem.
type FS interface {
	Close() error
	Cat(path string) ([]byte, error)
	Stat(path string) (os.FileInfo, error)
	LStat(path string) (os.FileInfo, error)
	Symlink(at, to string) error
	Mkdir(at string) error
	Write(path string, data []byte, perms os.FileMode) error
}

// FSMountProxy proxies access to a mounted filesystem.
type FSMountProxy struct {
	Kind     string
	Path     string
	fs       FS
	isClosed bool
}

// Close implements io.Closer.
func (p *FSMountProxy) Close() error {
	if p.isClosed {
		return nil
	}
	p.isClosed = true
	return p.fs.Close()
}

func (p *FSMountProxy) String() string {
	return fmt.Sprintf("fs.%sMount{%p}", p.Kind, p)
}

// Type implements starlark.Value.
func (p *FSMountProxy) Type() string {
	return fmt.Sprintf("fs.%sMount", p.Kind)
}

// Freeze implements starlark.Value.
func (p *FSMountProxy) Freeze() {
}

// Truth implements starlark.Value.
func (p *FSMountProxy) Truth() starlark.Bool {
	return starlark.Bool(true)
}

// Hash implements starlark.Value.
func (p *FSMountProxy) Hash() (uint32, error) {
	h := sha256.Sum256([]byte(p.String()))
	return uint32(uint32(h[0]) + uint32(h[1])<<8 + uint32(h[2])<<16 + uint32(h[3])<<24), nil
}

// AttrNames implements starlark.Value.
func (p *FSMountProxy) AttrNames() []string {
	return []string{"base", "cat"}
}

// Attr implements starlark.Value.
func (p *FSMountProxy) Attr(name string) (starlark.Value, error) {
	switch name {
	case "cat":
		return starlark.NewBuiltin("cat", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var path starlark.String
			if err := starlark.UnpackArgs("c", args, kwargs, "path", &path); err != nil {
				return starlark.None, err
			}

			d, err := p.fs.Cat(string(path))
			if err != nil {
				return starlark.None, err
			}
			return starlark.String(string(d)), nil
		}), nil
	case "base":
		return starlark.String(p.Path), nil
	}

	return nil, starlark.NoSuchAttrError(
		fmt.Sprintf("%s has no .%s attribute", p.Type(), name))
}
