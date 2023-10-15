package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"voicesummary/config"
	"voicesummary/storage"
	stt "voicesummary/stt"
)

func main() {
	// Define a string flag for the configuration file path
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")

	// Parse the flags
	flag.Parse()

	// Use the flag value
	config, err := config.GetConfigFromFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to get config: %v", err)
	}

	// Create a new Telegram bot
	bot, err := initializeBot(config.TelegramBotToken)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	ctx := context.Background()
	// Start processing updates
	stg, err := storage.NewRealStorage(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	speechToText, err := stt.NewGoogleSpeechToText(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create speechToText: %v", err)
	}

	processUpdatesLoop(ctx, bot, stg, speechToText)
}

func initializeBot(token string) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	bot.Debug = true
	if err == nil {
		log.Printf("Authorized on account %s", bot.Self.UserName)
	}
	return bot, err
}

func processUpdatesLoop(ctx context.Context, bot *tgbotapi.BotAPI, stg storage.Storage, stt stt.SpeechToText) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Failed to start getting updates: %v", err)
	}

	for update := range updates {
		log.Printf("%+v", update)
		if update.Message != nil && (update.Message.Audio != nil || update.Message.Voice != nil) {
			processMessage(ctx, update, bot, stg, stt)
		}
	}
}

// getAudioData takes a File object, downloads the associated audio file,
// and returns the file content as a byte slice or an error if one occurred.
func getAudioData(bot *tgbotapi.BotAPI, audioFile *tgbotapi.File) ([]byte, error) {
	// Create the download link for the file
	lnk := audioFile.Link(bot.Token)

	// Download the audio file.
	resp, err := http.Get(lnk)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio file: %v", err)
	}
	defer resp.Body.Close()

	// Read the audio file content.
	oggData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file content: %v", err)
	}

	return oggData, nil
}

func processMessage(ctx context.Context, update tgbotapi.Update, bot *tgbotapi.BotAPI, stg storage.Storage, speechToText stt.SpeechToText) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	audioFileID := getFileId(update)
	var file tgbotapi.File
	var err error
	file, err = bot.GetFile(tgbotapi.FileConfig{FileID: audioFileID})

	if err != nil {
		log.Printf("Failed to get audio file path: %+v", err)
		return
	}
	audioData, err := getAudioData(bot, &file)
	if err != nil {
		log.Printf("Failed to get audio data: %v", err)
		return
	}

	wavData, err := convertToWav(audioData)
	if err != nil {
		log.Printf("Failed to convert audio to WAV: %v", err)
		return
	}

	uri, err := stg.StoreFile(ctx, wavData)
	if err != nil {
		log.Printf("Failed to store audio file: %v", err)
		return
	}
	defer stg.ClearFile(ctx, uri)

	text, err := speechToText.RecognizeSpeechFromFile(ctx, uri)
	if err != nil {
		log.Printf("Failed to recognize speech: %v", err)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	bot.Send(msg)
}

// convertToWav converts raw audio bytes to WAV format.
func convertToWav(rawAudioBytes []byte) ([]byte, error) {
	// Create a temporary file to hold the raw audio data
	inp, err := os.CreateTemp("", "input-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(inp.Name())
	defer inp.Close()

	// Write the raw audio data to the temp file
	_, err = inp.Write(rawAudioBytes)
	if err != nil {
		return nil, err
	}

	// Create another temp file to store the converted WAV data
	wavFile, err := os.CreateTemp("", "output-*.wav")
	if err != nil {
		return nil, err
	}
	wavFileName := wavFile.Name()
	defer os.Remove(wavFileName)
	defer wavFile.Close()

	// Convert the file to WAV format using convertFileToWav
	err = convertFileToWav(inp, wavFileName, err)
	if err != nil {
		return nil, err
	}

	// Read the converted WAV data back into a byte slice
	wavData, err := os.ReadFile(wavFileName)
	if err != nil {
		return nil, err
	}

	return wavData, nil
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
