package main

import (
	"os"
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

func TestCleanUpTempFiles(t *testing.T) {
	tempFile, err := os.CreateTemp("", "comp2unraid-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	tempFiles = []string{tempFile.Name()}
	cleanUpTempFiles()
	if len(tempFiles) != 0 {
		t.Errorf("Expected tempFiles to be empty, but got %v", tempFiles)
	}
}

func TestConvertCommand(t *testing.T) {
	// This test is a bit more complex, as it requires a real config file and a real Docker Compose project.
	// You may need to modify this test to fit your specific use case.
}
