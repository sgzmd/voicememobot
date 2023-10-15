package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

// getTestFilePath returns the absolute path to the testdata file.
// It assumes that the file is in the same directory as the test.
func getTestFilePath(fileName string) string {
	_, currentFile, _, _ := runtime.Caller(1) // get the file name of the caller function
	dir := filepath.Dir(currentFile)          // get the directory of the current file
	return filepath.Join(dir, fileName)       // construct the path to the sample file
}

// TestGetConfigFromFile test scenarios for GetConfigFromFile function.
func TestGetConfigFromFile(t *testing.T) {
	t.Run("parse valid config", func(t *testing.T) {
		cfg, err := GetConfigFromFile(getTestFilePath("test_config.yaml"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cfg.TelegramBotToken != "sample_bot_token" {
			t.Errorf("expected %v, got %v", "sample_bot_token", cfg.TelegramBotToken)
		}
	})

	t.Run("handle non-existent file", func(t *testing.T) {
		_, err := GetConfigFromFile("path_to/non_existent_file.yaml")
		if err == nil {
			t.Fatalf("expected an error, got none")
		}
	})

	t.Run("handle config with missing fields", func(t *testing.T) {
		_, err := GetConfigFromFile("path_to/invalid_config.yaml")
		if err == nil {
			t.Fatalf("expected an error, got none")
		}
	})
}
