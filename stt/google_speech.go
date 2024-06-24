package stt

import (
	"context"
	"log"
	"sync"
	"time"
	"voicesummary/config"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
)

type GoogleSpeechToText struct {
	client *speech.Client
	once   sync.Once
	config *config.Config
}

func (r *GoogleSpeechToText) getSpeechClient(_ context.Context) *speech.Client {
	return r.client
}

func (r *GoogleSpeechToText) createRecognizerName() string {
	return "projects/" + r.config.GoogleSpeechProjectId +
		"/locations/global/recognizers/" + r.config.GoogleSpeechRecognizerName
}

func (r *GoogleSpeechToText) RecognizeSpeech(ctx context.Context, audio []byte) (string, error) {
	client := r.getSpeechClient(ctx)
	request := speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			LanguageCodes:  []string{"ru-RU", "en-GB"},
			Model:          "long",
			DecodingConfig: &speechpb.RecognitionConfig_AutoDecodingConfig{},
		},
		Recognizer: r.createRecognizerName(),
		AudioSource: &speechpb.RecognizeRequest_Content{
			Content: audio,
		},
	}

	resp, err := client.Recognize(ctx, &request)
	if err != nil {
		return "", err
	}

	log.Printf("Response: %+v", resp)
	if len(resp.Results) == 0 || len(resp.Results[0].Alternatives) == 0 {
		return "Could not transcribe text from the audio, empty audio clip?", nil
	}

	return resp.Results[0].Alternatives[0].Transcript, nil
}

func (r *GoogleSpeechToText) RecognizeSpeechFromFile(ctx context.Context, filePath string) (string, error) {
	client := r.getSpeechClient(ctx)
	request := r.createBatchRecognizeRequest(filePath)

	log.Printf("Request: %+v", request)

	op, err := client.BatchRecognize(ctx, &request)
	if err != nil {
		log.Printf("Failed to recognize stt: %+v", err)
		return "", err
	}

	resp, err := r.waitForRecognition(ctx, op)
	if err != nil {
		log.Printf("Failed to recognize stt: %+v", err)
		return "", err
	}

	result := extractTranscripts(resp, filePath)
	return result, nil
}

func (r *GoogleSpeechToText) createBatchRecognizeRequest(filePath string) speechpb.BatchRecognizeRequest {
	return speechpb.BatchRecognizeRequest{
		Recognizer: r.createRecognizerName(),
		Config: &speechpb.RecognitionConfig{
			LanguageCodes:  []string{"ru-RU", "en-GB"},
			Model:          "long",
			DecodingConfig: &speechpb.RecognitionConfig_AutoDecodingConfig{},
		},
		Files: []*speechpb.BatchRecognizeFileMetadata{
			{
				AudioSource: &speechpb.BatchRecognizeFileMetadata_Uri{
					Uri: filePath,
				},
			},
		},
		RecognitionOutputConfig: &speechpb.RecognitionOutputConfig{
			Output: &speechpb.RecognitionOutputConfig_InlineResponseConfig{},
		},
	}
}

func (r *GoogleSpeechToText) waitForRecognition(ctx context.Context, op *speech.BatchRecognizeOperation) (*speechpb.BatchRecognizeResponse, error) {
	var resp *speechpb.BatchRecognizeResponse
	var err error

	for {
		resp, err = op.Poll(ctx)
		if err != nil || op.Done() {
			break
		}

		log.Printf("Waiting for stt recognition operation to complete...")
		time.Sleep(1 * time.Second)
	}

	log.Printf("Response: %+v", resp)
	return resp, err
}

func extractTranscripts(resp *speechpb.BatchRecognizeResponse, filePath string) string {
	result := ""
	transcripts := resp.Results[filePath].Transcript.Results
	for _, transcript := range transcripts {
		if len(transcript.Alternatives) > 0 {
			result += transcript.Alternatives[0].Transcript + "\n"
		}
	}
	return result
}
