package stt

import (
	"context"
)

// FakeSpeechToText is a fake implementation of SpeechToText interface.
type FakeSpeechToText struct {
	RecognizeSpeechResponse         string
	RecognizeSpeechErr              error
	RecognizeSpeechFromFileResponse string
	RecognizeSpeechFromFileErr      error
}

// RecognizeSpeech returns fake recognition result for given audio input.
func (f *FakeSpeechToText) RecognizeSpeech(ctx context.Context, audio []byte) (string, error) {
	return f.RecognizeSpeechResponse, f.RecognizeSpeechErr
}

// RecognizeSpeechFromFile returns fake recognition result for given file path.
func (f *FakeSpeechToText) RecognizeSpeechFromFile(ctx context.Context, filePath string) (string, error) {
	return f.RecognizeSpeechFromFileResponse, f.RecognizeSpeechFromFileErr
}
