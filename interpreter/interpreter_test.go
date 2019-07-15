package interpreter

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/twitchyliquid64/raspberry-box/conf/sysd"
	"go.starlark.net/starlark"
)

var (
	imgPath = flag.String("pi-img", "", "Path to a mint raspbian image.")
)

func init() {
	flag.Parse()
}

func TestNewScript(t *testing.T) {
	var cVersion string
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		cVersion = args[0].String()
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`test_hook(compiler.version)`), "testNewScript.box", nil, nil, testCb)
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
test_hook(pi.library_version)`), "testNewScript.box", nil, nil, testCb)
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
			_, err := makeScript([]byte(code), "testMathBuiltins_"+tc.name+".box", nil, nil, testCb)
			if err != nil {
				t.Fatalf("makeScript() failed: %v", err)
			}

			if out != tc.expect {
				t.Errorf("test_hook() = %v, want %v", out, tc.expect)
			}
		})
	}
}

func TestScriptArgs(t *testing.T) {
	var a string
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		a = args[0].(starlark.String).GoString()
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`test_hook(args.arg(0) + " num=" + str(args.num_args()))`), "testNewScript.box", nil, []string{"--verbose", "test.img"}, testCb)
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if a != "test.img num=1" {
		t.Errorf("a = %v, want %v", a, "test.img num=1")
	}
}

func TestScriptFsPartitionsPiImage(t *testing.T) {
	if *imgPath == "" {
		t.SkipNow()
	}
	var a string
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		a = args[0].(starlark.String).GoString()
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`
test_hook(str(fs.read_partitions(args.arg(0))))`), "testScriptFsPartitionsPiImage.box", nil, []string{*imgPath}, testCb)
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	want := "[struct(bootable = False, empty = False, index = 0, lba = struct(length = 89854, start = 8192), type = 12, type_name = \"FAT32-LBA\"), struct(bootable = False, empty = False, index = 1, lba = struct(length = 3547136, start = 98304), type = 131, type_name = \"Native Linux\"), struct(bootable = False, empty = True, index = 2, lba = struct(length = 0, start = 0), type = 0, type_name = \"\"), struct(bootable = False, empty = True, index = 3, lba = struct(length = 0, start = 0), type = 0, type_name = \"\")]"
	if a != want {
		t.Errorf("a = %v, want %v", a, want)
	}
}

func TestBuildSysdUnit(t *testing.T) {
	var out starlark.Tuple
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		out = args
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`
unit = systemd.Unit(
	description="description",
	after=["kek", "startup.service"]
)
unit.description = "description yo"
unit.append_required_by(["woooo"], "mate")

serv = systemd.Service()
unit.service = serv

test_hook(unit, unit.description, unit.service)`), "testBuildSysdUnit.box", nil, nil, testCb)
	if err != nil {
		t.Fatalf("makeScript() failed: %v", err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if out[0].(*SystemdUnitProxy).Unit.Description != "description yo" {
		t.Errorf("out.Unit.Description = %v, want %v", out[0].(*SystemdUnitProxy).Unit.Description, "description yo")
	}
	if got, want := string(out[1].(starlark.String)), out[0].(*SystemdUnitProxy).Unit.Description; got != want {
		t.Errorf("unit.description = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdUnitProxy).Unit.After, []string{"kek", "startup.service"}; !reflect.DeepEqual(got, want) {
		t.Errorf("out.Unit.After = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdUnitProxy).Unit.RequiredBy, []string{"woooo", "mate"}; !reflect.DeepEqual(got, want) {
		t.Errorf("out.Unit.RequiredBy = %v, want %v", got, want)
	}
	if out[2].(*SystemdServiceProxy).Service != out[0].(*SystemdUnitProxy).Unit.Service {
		t.Errorf("out.Unit.Service (%v) != out.Service (%v)", out[0].(*SystemdUnitProxy).Unit.Service, out[2].(*SystemdServiceProxy).Service)
	}
}

func TestBuildSysdService(t *testing.T) {
	var out starlark.Tuple
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		out = args
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`
serv = systemd.Service(
	type="wrong",
	exec_start="echo kek",
	restart="always",
	restart_sec="15m",
	user="wrong",
	group="wrong",
	kill_mode="wooooo",
	stderr=88,
	stdout=99,
)
serv.set_user("also wrong")
serv.user = "root"
serv.set_group("also wrong")
serv.group = "root"
serv.set_type(systemd.const.service_simple)
serv.type = systemd.const.service_simple
serv.set_kill_mode(serv.kill_mode)
serv.kill_mode = systemd.const.killmode_controlgroup

serv.set_restart_sec("15s")
serv.restart_sec = serv.restart_sec
serv.set_timeout_stop_sec("10m15s")
serv.timeout_stop_sec = serv.timeout_stop_sec
serv.set_watchdog_sec("2h")
serv.watchdog_sec = serv.watchdog_sec

serv.set_ignore_sigpipe(False)
serv.ignore_sigpipe = True

serv.set_stdout(systemd.out.console)
serv.stdout = systemd.out.console + systemd.out.journal
serv.set_stderr(systemd.out.console)
serv.stderr = systemd.out.console + systemd.out.journal

test_hook(serv)`), "testBuildSysdService.box", nil, nil, testCb)
	if err != nil {
		t.Fatalf("makeScript() failed: %v", err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if got, want := out[0].(*SystemdServiceProxy).Service.Type, sysd.SimpleService; got != want {
		t.Errorf("out.Service.Type = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdServiceProxy).Service.ExecStart, "echo kek"; got != want {
		t.Errorf("out.Service.ExecStart = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdServiceProxy).Service.User, "root"; got != want {
		t.Errorf("out.Service.User = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdServiceProxy).Service.Group, "root"; got != want {
		t.Errorf("out.Service.Group = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdServiceProxy).Service.Restart, sysd.RestartAlways; got != want {
		t.Errorf("out.Service.Restart = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdServiceProxy).Service.RestartSec, time.Duration(15*time.Second); got != want {
		t.Errorf("out.Service.RestartSec = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdServiceProxy).Service.TimeoutStopSec, time.Duration(10*time.Minute+15*time.Second); got != want {
		t.Errorf("out.Service.TimeoutStopSec = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdServiceProxy).Service.WatchdogSec, time.Duration(2*time.Hour); got != want {
		t.Errorf("out.Service.WatchdogSec = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdServiceProxy).Service.KillMode, sysd.KMControlGroup; got != want {
		t.Errorf("out.Service.KillMode = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdServiceProxy).Service.IgnoreSigpipe, true; got != want {
		t.Errorf("out.Service.IgnoreSigpipe = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdServiceProxy).Service.Stdout, sysd.OutputConsole|sysd.OutputJournal; got != want {
		t.Errorf("out.Service.Stdout = %v, want %v", got, want)
	}
	if got, want := out[0].(*SystemdServiceProxy).Service.Stderr, sysd.OutputConsole|sysd.OutputJournal; got != want {
		t.Errorf("out.Service.Stderr = %v, want %v", got, want)
	}
}
