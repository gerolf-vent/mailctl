package utils

import "testing"

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		in   uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1 KB"},
		{1536, "1.5 KB"},
		{10 * 1024, "10 KB"},
		{1048576, "1 MB"},
		{5 * 1024 * 1024, "5 MB"},
		{123456789, "117.74 MB"},
		{1099511627776, "1 TB"},
	}

	for _, tc := range tests {
		got := FormatBytes(tc.in)
		if got != tc.want {
			t.Fatalf("FormatBytes(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
