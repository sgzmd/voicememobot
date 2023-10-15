package stt

import (
	speech "cloud.google.com/go/speech/apiv2"
	"context"
	"log"
	"sync"
	"voicesummary/config"
)

var client *speech.Client
var once sync.Once

type SpeechToText interface {
	RecognizeSpeech(ctx context.Context, audio []byte) (string, error)
	RecognizeSpeechFromFile(ctx context.Context, filePath string) (string, error)
}

func NewGoogleSpeechToText(ctx context.Context, cfg *config.Config) (*GoogleSpeechToText, error) {
	client, err := speech.NewClient(ctx, cfg.GetCredentialsOption())
	if err != nil {
		log.Printf("Failed to create speech client: %v", err)
		return nil, err
	}

	return &GoogleSpeechToText{
		client: client,
		config: cfg,
	}, nil
}
