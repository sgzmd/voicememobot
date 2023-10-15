package stt

import (
	"cloud.google.com/go/speech/apiv2/speechpb"
	"testing"
	"voicesummary/config"
)

func TestCreateRecognizerName(t *testing.T) {
	tests := []struct {
		name               string
		projectID          string
		recognizerName     string
		expectedRecognizer string
	}{
		{
			name:               "Valid IDs",
			projectID:          "project123",
			recognizerName:     "recognizerABC",
			expectedRecognizer: "projects/project123/locations/global/recognizers/recognizerABC",
		},
		{
			name:               "Empty ProjectID",
			projectID:          "",
			recognizerName:     "recognizerABC",
			expectedRecognizer: "projects//locations/global/recognizers/recognizerABC",
		},
		{
			name:               "Empty RecognizerName",
			projectID:          "project123",
			recognizerName:     "",
			expectedRecognizer: "projects/project123/locations/global/recognizers/",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{
				GoogleSpeechProjectId:      test.projectID,
				GoogleSpeechRecognizerName: test.recognizerName,
			}
			r := &GoogleSpeechToText{config: cfg}

			got := r.createRecognizerName()
			if got != test.expectedRecognizer {
				t.Errorf("Expected recognizer name: %v, got: %v", test.expectedRecognizer, got)
			}
		})
	}
}

func TestCreateBatchRecognizeRequest(t *testing.T) {
	r := &GoogleSpeechToText{
		config: &config.Config{
			GoogleSpeechProjectId:      "project123",
			GoogleSpeechRecognizerName: "recognizerABC",
		},
	}

	filePath := "path/to/file"
	request := r.createBatchRecognizeRequest(filePath)

	if got, want := request.Recognizer, "projects/project123/locations/global/recognizers/recognizerABC"; got != want {
		t.Errorf("Incorrect recognizer. Got %v, want %v", got, want)
	}

	audioSourceUri, ok := request.Files[0].AudioSource.(*speechpb.BatchRecognizeFileMetadata_Uri)
	if !ok {
		t.Fatal("AudioSource is not of type *BatchRecognizeFileMetadata_Uri")
	}

	if got, want := audioSourceUri.Uri, filePath; got != want {
		t.Errorf("Incorrect file URI. Got %v, want %v", got, want)
	}
}
