package interpreter

import (
	"errors"
	"fmt"
	"testing"

	"go.starlark.net/starlark"
)

func TestNewScript(t *testing.T) {
	var cVersion string
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		cVersion = args[0].String()
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`test_hook(compiler.version)`), "testNewScript.box", nil, testCb)
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if cVersion != fmt.Sprint(starlark.CompilerVersion) {
		t.Errorf("cVersion = %v, want %v", cVersion, starlark.CompilerVersion)
	}
}

func TestLoadScript(t *testing.T) {
	var cVersion string
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		cVersion = args[0].String()
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`load("pi.lib", "pi")
test_hook(pi.library_version)`), "testNewScript.box", nil, testCb)
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if cVersion != "1" {
		t.Errorf("cVersion = %v, want %v", cVersion, 1)
	}
}

func TestMathBuiltins(t *testing.T) {
	tcs := []struct {
		name   string
		a1, a2 uint64
		method string
		expect uint64
	}{
		{
			name:   "and 1",
			method: "_and",
			a1:     1,
			a2:     3,
			expect: 1,
		},
		{
			name:   "and 2",
			method: "_and",
			a1:     1,
			a2:     0,
			expect: 0,
		},
		{
			name:   "and 3",
			method: "_and",
			a1:     15,
			a2:     4,
			expect: 4,
		},
		{
			name:   "shl 1",
			method: "shl",
			a1:     1,
			a2:     1,
			expect: 2,
		},
		{
			name:   "shl 2",
			method: "shl",
			a1:     1,
			a2:     2,
			expect: 4,
		},
		{
			name:   "shl 3",
			method: "shl",
			a1:     16,
			a2:     2,
			expect: 64,
		},
		{
			name:   "shr 1",
			method: "shr",
			a1:     16,
			a2:     1,
			expect: 8,
		},
		{
			name:   "shr 2",
			method: "shr",
			a1:     2,
			a2:     6,
			expect: 0,
		},
		{
			name:   "shr 3",
			method: "shr",
			a1:     15,
			a2:     3,
			expect: 1,
		},
		{
			name:   "not 1",
			method: "_not",
			a1:     1,
			a2:     9999999999,
			expect: 18446744073709551614,
		},
		{
			name:   "not 2",
			method: "_not",
			a1:     18446744073709551610,
			a2:     9999999999,
			expect: 5,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var out uint64
			testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
				var result starlark.Int
				if err := starlark.UnpackArgs("test_hook", args, kwargs, "result", &result); err != nil {
					return starlark.None, err
				}
				b, ok := result.Uint64()
				if !ok {
					return starlark.None, errors.New("cannot represent result as unsigned integer")
				}
				out = b
				return starlark.None, nil
			}

			code := fmt.Sprintf("test_hook(math.%s(%d", tc.method, tc.a1)
			if tc.a2 != 9999999999 {
				code += fmt.Sprintf(", %d))", tc.a2)
			} else {
				code += "))"
			}

			t.Logf("code = %q", code)
			_, err := makeScript([]byte(code), "testMathBuiltins_"+tc.name+".box", nil, testCb)
			if err != nil {
				t.Fatalf("makeScript() failed: %v", err)
			}

			if out != tc.expect {
				t.Errorf("test_hook() = %v, want %v", out, tc.expect)
			}
		})
	}
}
