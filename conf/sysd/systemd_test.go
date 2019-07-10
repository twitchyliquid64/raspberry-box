package sysd

import (
	"testing"
	"time"
)

func TestSystemdService(t *testing.T) {
	tcs := []struct {
		name string
		inp  Service
		out  string
	}{
		{
			name: "empty",
			out:  "[Service]\nIgnoreSIGPIPE=no\n",
		},
		{
			name: "basic",
			inp: Service{
				ExecStart: "/bin/echo yolo",
				Stdout:    OutputConsole,
			},
			out: "[Service]\nExecStart=/bin/echo yolo\nIgnoreSIGPIPE=no\nStandardOutput=console\n",
		},
		{
			name: "restart",
			inp: Service{
				Restart:    RestartAlways,
				RestartSec: 5 * time.Second,
				Stdout:     OutputConsole,
			},
			out: "[Service]\nRestart=always\nRestartSec=5s\nIgnoreSIGPIPE=no\nStandardOutput=console\n",
		},
		{
			name: "complex",
			inp: Service{
				NotifyAccess: NotifyAllProcs,
				Stdout:       OutputConsole | OutputJournal,
				Type:         NotifyService,
				KillMode:     KMControlGroup,
				Restart:      RestartNever,
			},
			out: "[Service]\nType=notify\nKillMode=control-group\nRestart=no\nNotifyAccess=all\nIgnoreSIGPIPE=no\nStandardOutput=console\n",
		},
		{
			name: "conditions",
			inp: Service{
				ExecStart: "echo yolo swaggins",
				Conditions: Conditions{
					ConditionExists("/bin/echo"),
					ConditionHost("pi2"),
				},
			},
			out: "[Service]\nExecStart=echo yolo swaggins\nIgnoreSIGPIPE=no\nConditionPathExists=/bin/echo\nConditionHost=pi2\n",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.out != tc.inp.String() {
				t.Errorf("out = %q, want %q", tc.inp.String(), tc.out)
			}
		})
	}
}

func TestSystemdUnit(t *testing.T) {
	tcs := []struct {
		name string
		inp  Unit
		out  string
	}{
		{
			name: "empty",
			out:  "[Unit]\n\n",
		},
		{
			name: "basic",
			inp: Unit{
				Description: "yolo",
			},
			out: "[Unit]\nDescription=yolo\n\n",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.out != tc.inp.String() {
				t.Errorf("out = %q, want %q", tc.inp.String(), tc.out)
			}
		})
	}
}
