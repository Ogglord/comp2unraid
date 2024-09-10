package main

import (
	"testing"
)

func TestGetRegistryURL(t *testing.T) {
	tests := []struct {
		image string
		want  string
	}{
		{"quay.io/nextcloud/server", "https://quay.io/repository/nextcloud/server"},
		{"ghcr.io/nextcloud/server", "https://github.com/nextcloud/server"},
		{"docker.io/nextcloud/server", "https://hub.docker.com/r/nextcloud/server"},
	}

	for _, tt := range tests {
		got, err := getRegistryURL(tt.image)
		if err != nil {
			t.Errorf("getRegistryURL(%q) error = %v", tt.image, err)
		}
		if got != tt.want {
			t.Errorf("getRegistryURL(%q) = %q, want %q", tt.image, got, tt.want)
		}
	}
}

func TestConvertCommand(t *testing.T) {
	// This test is a bit more complex, as it requires a real config file and a real Docker Compose project.
	// You may need to modify this test to fit your specific use case.
}
