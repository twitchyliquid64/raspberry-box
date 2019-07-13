// Package interpreter implements the execution of raspberry-box box files.
package interpreter

import (
	"errors"
	"fmt"

	"github.com/twitchyliquid64/raspberry-box/interpreter/lib"
	"go.starlark.net/starlark"
)

// Script represents a raspberry-box script.
type Script struct {
	loader ScriptLoader

	thread  *starlark.Thread
	globals starlark.StringDict

	// testHook is only accessible and populated from unit tests.
	testHook func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)
}

// NewScript initializes a new raspberry-box script environment.
func NewScript(data []byte, fname string, loader ScriptLoader) (*Script, error) {
	return makeScript(data, fname, loader, nil)
}

func makeScript(data []byte, fname string, loader ScriptLoader, testHook func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)) (*Script, error) {
	out := &Script{
		loader:   loader,
		testHook: testHook,
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
