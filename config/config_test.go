package config

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func getTestFilePath(fileName string) string {
	_, currentFile, _, _ := runtime.Caller(1)
	dir := filepath.Dir(currentFile)
	return filepath.Join(dir, fileName)
}

func TestGetConfigFromFile(t *testing.T) {
	t.Run("parse valid config", func(t *testing.T) {
		cfg, err := GetConfigFromFile(getTestFilePath("test_config.yaml"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg.TelegramBotToken != "sample_bot_token" {
			t.Errorf("expected %v, got %v", "sample_bot_token", cfg.TelegramBotToken)
		}

		expectedUsernames := []string{"sample_username", "sample_username2"}
		if !reflect.DeepEqual(cfg.Usernames, expectedUsernames) {
			t.Errorf("expected usernames %v, got %v", expectedUsernames, cfg.Usernames)
		}
	})

	t.Run("handle non-existent file", func(t *testing.T) {
		_, err := GetConfigFromFile("path_to/non_existent_file.yaml")
		if err == nil {
			t.Fatalf("expected an error, got none")
		}
	})

	t.Run("handle config with missing fields", func(t *testing.T) {
		_, err := GetConfigFromFile(getTestFilePath("invalid_config.yaml"))
		if err == nil {
			t.Fatalf("expected an error, got none")
		}
	})
}
