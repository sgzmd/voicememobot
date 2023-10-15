package main

import (
	"context"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"voicesummary/config"
	"voicesummary/storage"
	stt "voicesummary/stt"
)

func main() {
	config := config.GetConfigFromEnv()

	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	bot.Debug = true
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Failed to start getting updates: %v", err)
	}

	for update := range updates {
		log.Printf("%+v", update)
		if update.Message != nil && (update.Message.Audio != nil || update.Message.Voice != nil) {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			// Get the audio file path.
			audioFile := getFileId(update)
			file, err := bot.GetFile(tgbotapi.FileConfig{FileID: audioFile})

			if err != nil {
				log.Printf("Failed to get audio file path: %+v", err)
				continue
			}

			lnk := file.Link(bot.Token)
			log.Print(file.FilePath)

			// Download the audio file.
			resp, err := http.Get(lnk)
			if err != nil {
				log.Printf("Failed to download audio file: %v", err)
				continue
			}
			defer resp.Body.Close()

			// Read the audio file content.
			oggData, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read audio file content: %v", err)
				continue
			}

			tempdir := os.TempDir()
			inp, err := os.CreateTemp(tempdir, "input-")
			if err != nil {
				log.Printf("Failed to create temp file: %v", err)
				continue
			}
			defer inp.Close()
			_, err = inp.Write(oggData)

			wavFile, err := os.CreateTemp(tempdir, "wav-*.wav")
			if err != nil {
				log.Printf("Failed to create temp file: %v", err)
				continue
			}
			wavFileName := wavFile.Name()
			defer os.Remove(wavFileName)
			defer wavFile.Close()

			err = convertFileToWav(inp, wavFileName, err)
			if err != nil {
				log.Printf("Failed to convert file: %+v", err)
				continue
			}

			// read wavfile into rawAudio
			rawAudio, err := io.ReadAll(wavFile)
			if err != nil {
				log.Printf("Failed to read audio file content: %v", err)
				continue
			}

			uri, err := storage.StoreFile(config.GoogleStorageBucket, rawAudio)
			if err != nil {
				log.Printf("Failed to store audio file: %v", err)
				continue
			} else {
				log.Printf("Stored audio file: %v", uri)
			}
			defer storage.ClearFile(config.GoogleStorageBucket, uri)

			text, err := stt.RecognizeSpeechBucketFile(context.Background(), uri)
			if err != nil {
				log.Printf("Failed to recognize stt: %+v", err)
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			bot.Send(msg)
		}
	}
}

func processUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message == nil || (update.Message.Audio == nil && update.Message.Voice == nil) {
		return
	}

	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
}

// convertFileToWav converts an audio file in some format to wav format, 16kHz, mono.
func convertFileToWav(inp *os.File, wavFileName string, err error) error {
	cmd := exec.Command("ffmpeg", "-i", inp.Name(), "-acodec", "pcm_s16le", "-ac", "1", "-ar", "16000", "-y", wavFileName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	return err
}

// getFileId returns file ID of either voice memo or audio file sent by the user,
// depending on which one is present.
func getFileId(update tgbotapi.Update) string {
	if update.Message.Voice != nil {
		return update.Message.Voice.FileID
	}
	return update.Message.Audio.FileID
}
