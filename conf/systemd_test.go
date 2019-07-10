package conf

import (
	"testing"
	"time"
)

func TestSystemdService(t *testing.T) {
	tcs := []struct {
		name string
		inp  SystemdService
		out  string
	}{
		{
			name: "empty",
			out:  "[Service]\nIgnoreSIGPIPE=no\n",
		},
		{
			name: "basic",
			inp: SystemdService{
				ExecStart: "/bin/echo yolo",
				Stdout:    OutputConsole,
			},
			out: "[Service]\nExecStart=/bin/echo yolo\nIgnoreSIGPIPE=no\nStandardOutput=console\n",
		},
		{
			name: "restart",
			inp: SystemdService{
				Restart:    RestartAlways,
				RestartSec: 5 * time.Second,
				Stdout:     OutputConsole,
			},
			out: "[Service]\nRestart=always\nRestartSec=5s\nIgnoreSIGPIPE=no\nStandardOutput=console\n",
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
		inp  SystemdUnit
		out  string
	}{
		{
			name: "empty",
			out:  "[Unit]\n\n",
		},
		{
			name: "basic",
			inp: SystemdUnit{
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
