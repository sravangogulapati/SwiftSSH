package main

import "testing"

func TestExtractConfigFlag(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"--config", "/tmp/myconfig"}, "/tmp/myconfig"},
		{[]string{"-config", "/tmp/myconfig"}, "/tmp/myconfig"},
		{[]string{"--config=/etc/ssh/config"}, "/etc/ssh/config"},
		{[]string{"-config=/etc/ssh/config"}, "/etc/ssh/config"},
		{[]string{"--no-frequent"}, ""},
		{[]string{}, ""},
		{[]string{"--config"}, ""}, // missing value
	}
	for _, tc := range tests {
		if got := extractConfigFlag(tc.args); got != tc.want {
			t.Errorf("extractConfigFlag(%v) = %q; want %q", tc.args, got, tc.want)
		}
	}
}
