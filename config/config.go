package config

import (
	"errors"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"strings"
)

const (
	telegramBotTokenKey        = "TELEGRAM_BOT_TOKEN"
	googleSpeechApiFile        = "GOOGLE_SPEECH_API_FILE"
	googleSpeechProjectId      = "GOOGLE_SPEECH_PROJECT_ID"
	googleSpeechRecognizerName = "GOOGLE_SPEECH_RECOGNIZER_NAME"
	googleStorageBucket        = "GOOGLE_STORAGE_BUCKET"
)

type Config struct {
	TelegramBotToken           string `yaml:"telegram_bot_token"`
	GoogleSpeechApiKey         string `yaml:"google_speech_api_key"`
	GoogleSpeechProjectId      string `yaml:"google_speech_project_id"`
	GoogleSpeechRecognizerName string `yaml:"google_speech_recognizer_name"`
	GoogleStorageBucket        string `yaml:"google_storage_bucket"`
}

func (c *Config) GetCredentialsOption() option.ClientOption {
	return option.WithCredentialsFile(c.GoogleSpeechApiKey)
}

func GetConfigFromEnv() *Config {
	cfg := Config{
		TelegramBotToken:           os.Getenv(telegramBotTokenKey),
		GoogleSpeechApiKey:         os.Getenv(googleSpeechApiFile),
		GoogleSpeechProjectId:      os.Getenv(googleSpeechProjectId),
		GoogleSpeechRecognizerName: os.Getenv(googleSpeechRecognizerName),
		GoogleStorageBucket:        os.Getenv(googleStorageBucket),
	}

	if cfg.TelegramBotToken == "" ||
		cfg.GoogleSpeechProjectId == "" ||
		cfg.GoogleSpeechRecognizerName == "" ||
		cfg.GoogleSpeechApiKey == "" {
		log.Fatalf("Failed to get config: %+v", cfg)
	}

	return &cfg
}

// GetConfigFromFile reads the YAML file from the provided path,
// unmarshals it into a Config struct, and validates the fields.
func GetConfigFromFile(filePath string) (*Config, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the YAML data into a Config struct
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	// Validate the config fields
	err = validateConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validateConfig checks if all fields of the Config struct are non-empty.
func validateConfig(cfg Config) error {
	var missingFields []string

	if cfg.TelegramBotToken == "" {
		missingFields = append(missingFields, "TelegramBotToken")
	}
	if cfg.GoogleSpeechApiKey == "" {
		missingFields = append(missingFields, "GoogleSpeechApiKey")
	}
	if cfg.GoogleSpeechProjectId == "" {
		missingFields = append(missingFields, "GoogleSpeechProjectId")
	}
	if cfg.GoogleSpeechRecognizerName == "" {
		missingFields = append(missingFields, "GoogleSpeechRecognizerName")
	}
	if cfg.GoogleStorageBucket == "" {
		missingFields = append(missingFields, "GoogleStorageBucket")
	}

	if len(missingFields) > 0 {
		return errors.New("missing mandatory fields: " + strings.Join(missingFields, ", "))
	}

	return nil
}
