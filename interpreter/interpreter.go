// Package interpreter implements the execution of raspberry-box box files.
package interpreter

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/twitchyliquid64/raspberry-box/interpreter/lib"
	"go.starlark.net/starlark"
)

// Script represents a raspberry-box script.
type Script struct {
	loader ScriptLoader

	args    []string
	fs      *flag.FlagSet
	verbose bool

	thread  *starlark.Thread
	globals starlark.StringDict

	resources []io.Closer

	// testHook is only accessible and populated from unit tests.
	testHook func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)
}

// Close shuts down all resources associated with the script.
func (s *Script) Close() error {
	for _, r := range s.resources {
		r.Close()
	}
	return nil
}

// NewScript initializes a new raspberry-box script environment.
func NewScript(data []byte, fname string, loader ScriptLoader, args []string) (*Script, error) {
	return makeScript(data, fname, loader, args, nil)
}

func makeScript(data []byte, fname string, loader ScriptLoader, args []string,
	testHook func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)) (*Script, error) {
	out := &Script{
		loader:   loader,
		testHook: testHook,
		args:     args,
	}

	if err := out.initFlags(); err != nil {
		return nil, err
	}

	var err error
	out.thread, out.globals, err = out.loadScript(data, fname, out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Script) printFromSkylark(_ *starlark.Thread, msg string) {
	fmt.Println(msg)
}

func (s *Script) resolveImport(path string) ([]byte, error) {
	d, exists := lib.Libs[path]
	if exists {
		return d, nil
	}
	if s.loader == nil {
		return nil, errors.New("no such import: " + path)
	}
	return s.loader.resolveImport(path)
}

func cvStrListToStarlark(in []string) *starlark.List {
	out := make([]starlark.Value, len(in))
	for i := range in {
		out[i] = starlark.String(in[i])
	}
	return starlark.NewList(out)
}
