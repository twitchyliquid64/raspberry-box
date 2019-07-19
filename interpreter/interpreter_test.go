package interpreter

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	cnet "github.com/twitchyliquid64/raspberry-box/conf/net"
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

	s, err := makeScript([]byte(`test_hook(compiler.version)`), "testNewScript.box", nil, nil, false, testCb)
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
test_hook(pi.library_version)`), "testNewScript.box", nil, nil, false, testCb)
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if cVersion != "2" {
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

			_, err := makeScript([]byte(code), "testMathBuiltins_"+tc.name+".box", nil, nil, false, testCb)
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

	s, err := makeScript([]byte(`test_hook(args.arg(0) + " num=" + str(args.num_args()))`), "testNewScript.box", nil, []string{"test.img"}, false, testCb)
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
test_hook(str(fs.read_partitions(args.arg(0))))`), "testScriptFsPartitionsPiImage.box", nil, []string{*imgPath}, false, testCb)
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

test_hook(unit, unit.description, unit.service)`), "testBuildSysdUnit.box", nil, nil, false, testCb)
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
	exec_stop="echo ending",
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

serv.exec_reload = serv.exec_reload
serv.set_exec_reload('aaa')
serv.exec_stop = serv.exec_stop
serv.set_exec_stop(serv.exec_stop)
serv.exec_start_pre = serv.exec_start_pre
serv.set_exec_start_pre(serv.exec_start_pre)
serv.exec_stop_post = serv.exec_stop_post
serv.set_exec_stop_post(serv.exec_stop_post)

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

serv.set_conditions([systemd.ConditionExists("/bin/systemd"), systemd.ConditionHost("aaa")])
serv.conditions = [systemd.ConditionNotExists("/bin/systemd"), serv.conditions[0], serv.conditions[1]]

test_hook(serv)`), "testBuildSysdService.box", nil, nil, false, testCb)
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
	if got, want := out[0].(*SystemdServiceProxy).Service.Conditions, (sysd.Conditions{
		sysd.ConditionNotExists("/bin/systemd"),
		sysd.ConditionExists("/bin/systemd"),
		sysd.ConditionHost("aaa"),
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("out.Service.Conditions = %v, want %v", got, want)
	}
}

func TestBuildSysdCondition(t *testing.T) {
	var out starlark.Tuple
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		out = args
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`
c = systemd.ConditionHost("aaaa")
test_hook(c, c.arg)`), "testBuildSysdCondition.box", nil, nil, false, testCb)
	if err != nil {
		t.Fatalf("makeScript() failed: %v", err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if got, want := out[0].(*SystemdConditionProxy).Kind, "ConditionHost"; got != want {
		t.Errorf("out.Kind = %v, want %v", got, want)
	}
	if got, want := string(out[1].(starlark.String)), "aaaa"; got != want {
		t.Errorf("out[1] = %v, want %v", got, want)
	}
}

func TestBuildNetDHCPProfile(t *testing.T) {
	var out starlark.Tuple
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		out = args
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`
c = net.DHCPProfile(
	name = "dhcp test",
	lease_seconds=10,
	hostname="kek",
	dns=True,
)
c.set_lease_seconds(25)
c.set_client_id(True)
test_hook(c, c.name)`), "testBuildNetDHCPProfile.box", nil, nil, false, testCb)
	if err != nil {
		t.Fatalf("makeScript() failed: %v", err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if got, want := out[0].(*DHCPProfileProxy).Profile.DHCP.LeaseSeconds, 25; got != want {
		t.Errorf("out.DHCP.LeaseSeconds = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPProfileProxy).Profile.DHCP.Hostname, "kek"; got != want {
		t.Errorf("out.DHCP.Hostname = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPProfileProxy).Profile.DHCP.PresentHostname, true; got != want {
		t.Errorf("out.DHCP.PresentHostname = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPProfileProxy).Profile.DHCP.ClientID, true; got != want {
		t.Errorf("out.DHCP.ClientID = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPProfileProxy).Profile.DHCP.SetupDNS, true; got != want {
		t.Errorf("out.DHCP.SetupDNS = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPProfileProxy).Profile.DHCP.RequestNTP, false; got != want {
		t.Errorf("out.DHCP.RequestNTP = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPProfileProxy).Profile.ProfileName, string(out[1].(starlark.String)); got != want {
		t.Errorf("out.ProfileName = %v, want %v", got, want)
	}
}

func TestBuildNetStaticProfile(t *testing.T) {
	var out starlark.Tuple
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		out = args
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`
c = net.StaticProfile(
	interface = "eth0",
	network = "192.168.1.77/24",
	broadcast = "192.168.1.111",
	routers = ['192.168.1.55'],
	dns = ['1.1.1.1'],
)
c.set_network(c.network)
c.network = '192.168.1.5/24'
c.set_broadcast("192.168.1.255")
c.set_routers(c.routers)
c.routers = ['192.168.1.1']
c.set_dns(c.dns)
c.dns = ['8.8.8.8']
test_hook(c, c.name)`), "testBuildNetStaticProfile.box", nil, nil, false, testCb)
	if err != nil {
		t.Fatalf("makeScript() failed: %v", err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if got, want := out[0].(*DHCPProfileProxy).Profile.InterfaceName, "eth0"; got != want {
		t.Errorf("out.InterfaceName = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPProfileProxy).Profile.Static.Network, (net.IPNet{
		IP:   net.ParseIP("192.168.1.5"),
		Mask: net.CIDRMask(24, 32),
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("out.Static.Network = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPProfileProxy).Profile.Static.Broadcast, net.ParseIP("192.168.1.255"); !reflect.DeepEqual(got, want) {
		t.Errorf("out.Static.Broadcast = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPProfileProxy).Profile.Static.Routers, []net.IP{net.ParseIP("192.168.1.1")}; !reflect.DeepEqual(got, want) {
		t.Errorf("out.Static.Routers = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPProfileProxy).Profile.Static.DNS, []string{"8.8.8.8"}; !reflect.DeepEqual(got, want) {
		t.Errorf("out.Static.DNS = %v, want %v", got, want)
	}
}

func TestBuildNetDHCPCD(t *testing.T) {
	var out starlark.Tuple
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		out = args
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`
s = net.StaticProfile(
	interface = "eth0",
	network = "192.168.1.5/24",
	broadcast = "192.168.1.111",
	routers = ['192.168.1.1'],
	dns = ['8.8.8.8'],
)

c = net.DHCPClient(
	control_group='wheelie',
	profiles = [s],
)

c.set_control_group(c.control_group[:5])
c.control_group = c.control_group

d = net.DHCPProfile(interface='wlan0', hostname='pi1')
c.profiles = [c.profiles[0], d]

test_hook(c)`), "testBuildNetDHCPCD.box", nil, nil, false, testCb)
	if err != nil {
		t.Fatalf("makeScript() failed: %v", err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if got, want := out[0].(*DHCPClientConfProxy).Conf.ControlGroup, "wheel"; got != want {
		t.Errorf("out.ControlGroup = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPClientConfProxy).Conf.Sections[0].InterfaceName, "eth0"; got != want {
		t.Errorf("out.Sections[0].InterfaceName = %v, want %v", got, want)
	}
	if got, want := out[0].(*DHCPClientConfProxy).Conf.Sections[1].InterfaceName, "wlan0"; got != want {
		t.Errorf("out.Sections[0].InterfaceName = %v, want %v", got, want)
	}
}

func TestBuildNetWifiNetwork(t *testing.T) {
	var out starlark.Tuple
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		out = args
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`
n = net.wifi.Network(
	ssid = 'test',
	psk = 'woooooo',
)
n.set_ssid(n.ssid)
n.ssid = n.ssid + ' network'
n.set_disabled(not n.disabled)
n.disabled = n.disabled
n.set_psk(n.psk)
n.psk = n.psk
test_hook(n)`), "testBuildNetWifiNetwork.box", nil, nil, false, testCb)
	if err != nil {
		t.Fatalf("makeScript() failed: %v", err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if got, want := int(out[0].(*WifiNetworkProxy).Conf.Mode), 0; got != want {
		t.Errorf("out.Mode = %v, want %v", got, want)
	}
	if got, want := out[0].(*WifiNetworkProxy).Conf.Disabled, true; got != want {
		t.Errorf("out.Disabled = %v, want %v", got, want)
	}
	if got, want := out[0].(*WifiNetworkProxy).Conf.SSID, "test network"; got != want {
		t.Errorf("out.SSID = %v, want %v", got, want)
	}
	if got, want := out[0].(*WifiNetworkProxy).Conf.PSK, "woooooo"; got != want {
		t.Errorf("out.PSK = %v, want %v", got, want)
	}
}

func TestBuildNetWifiConfig(t *testing.T) {
	var out starlark.Tuple
	testCb := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		out = args
		return starlark.None, nil
	}

	s, err := makeScript([]byte(`
n1 = net.wifi.Network(
	ssid = 'test',
	psk = 'woooooo',
)

c = net.wifi.SupplicantConfig(
	control_interface='/run/wpa_supplicant',
	allow_update_config = True,
	country_code='US',
	device_name='wlan0',
	networks=[n1],
)

n2 = net.wifi.Network(
	ssid = 'nope',
	disabled = True,
	psk = 'aaa',
)
c.networks = c.networks
c.set_networks([c.networks[0], n2])

c.set_control_interface(c.control_interface)
c.control_interface = c.control_interface
c.set_allow_update_config(not c.allow_update_config)
c.allow_update_config = not c.allow_update_config
c.country_code = c.country_code + 'aaaa'
c.set_country_code(c.country_code[0:2])
c.device_name = c.device_name + '121'
c.set_device_name(c.device_name[0:5])

test_hook(c)`), "testBuildNetWifiConfig.box", nil, nil, false, testCb)
	if err != nil {
		t.Fatalf("makeScript() failed: %v", err)
	}
	if s == nil {
		t.Error("script is nil")
	}

	if got, want := out[0].(*WifiConfigurationProxy).Conf.CtrlInterface, "/run/wpa_supplicant"; got != want {
		t.Errorf("out.CtrlInterface = %v, want %v", got, want)
	}
	if got, want := out[0].(*WifiConfigurationProxy).Conf.CountryCode, "US"; got != want {
		t.Errorf("out.CountryCode = %v, want %v", got, want)
	}
	if got, want := out[0].(*WifiConfigurationProxy).Conf.AllowUpdateConfig, true; got != want {
		t.Errorf("out.AllowUpdateConfig = %v, want %v", got, want)
	}
	if got, want := out[0].(*WifiConfigurationProxy).Conf.DeviceName, "wlan0"; got != want {
		t.Errorf("out.DeviceName = %v, want %v", got, want)
	}
	if got, want := out[0].(*WifiConfigurationProxy).Conf.Networks, []cnet.WPASupplicantNetwork{
		{
			SSID: "test",
			PSK:  "woooooo",
		},
		{
			SSID:     "nope",
			Disabled: true,
			PSK:      "aaa",
		},
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("out.Networks = %v, want %v", got, want)
	}
}
