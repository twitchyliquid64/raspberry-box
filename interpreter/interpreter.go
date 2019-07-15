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

	thread   *starlark.Thread
	globals  starlark.StringDict
	setupVal starlark.Value

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
func NewScript(data []byte, fname string, verbose bool, loader ScriptLoader, args []string) (*Script, error) {
	return makeScript(data, fname, loader, args, verbose, nil)
}

func makeScript(data []byte, fname string, loader ScriptLoader, args []string, verbose bool,
	testHook func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)) (*Script, error) {
	out := &Script{
		loader:   loader,
		testHook: testHook,
		args:     args,
		verbose:  verbose,
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

// Setup calls the setup() method in the script.
func (s *Script) Setup(templatePath string) error {
	if fn, exists := s.globals["setup"]; exists {
		setupVal, err := starlark.Call(s.thread, fn, starlark.Tuple{starlark.String(templatePath)}, nil)
		if err != nil {
			return err
		}
		s.setupVal = setupVal
	} else {
		s.setupVal = starlark.None
	}
	return nil
}

// Build calls the build() method in the script.
func (s *Script) Build() error {
	fn, exists := s.globals["build"]
	if !exists {
		return errors.New("build() function not present")
	}
	if _, err := starlark.Call(s.thread, fn, starlark.Tuple{s.setupVal}, nil); err != nil {
		return err
	}
	return nil
}

// CallFn calls an arbitrary function
func (s *Script) CallFn(fname string) (string, error) {
	fn, exists := s.globals[fname]
	if !exists {
		return "", fmt.Errorf("%s() function not present", fname)
	}
	ret, err := starlark.Call(s.thread, fn, starlark.Tuple{}, nil)
	if err != nil {
		return "", err
	}
	result, ok := ret.(starlark.String)
	if !ok {
		return "", fmt.Errorf("%s() returned type %T, want string", fname, ret)
	}
	return string(result), nil
}
