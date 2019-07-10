package net

import (
	"testing"
)

func TestWPASupplicant(t *testing.T) {
	tcs := []struct {
		name string
		inp  WPASupplicantConfig
		out  string
	}{
		{
			name: "empty",
			out:  "update_config=0\n",
		},
		{
			name: "simple",
			inp: WPASupplicantConfig{
				CountryCode:       "US",
				AllowUpdateConfig: true,
			},
			out: "update_config=1\ncountry=US\n",
		},
		{
			name: "single psk",
			inp: WPASupplicantConfig{
				CountryCode:       "US",
				AllowUpdateConfig: true,
				Networks: []WPASupplicantNetwork{
					{
						Mode: ModeClient,
						SSID: "yolo",
						PSK:  "swaggins",
					},
				},
			},
			out: "update_config=1\ncountry=US\nnetwork={\n\tmode=0\n\tdisabled=0\n\tssid=\"yolo\"\n\tpsk=\"swaggins\"\n}\n",
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
