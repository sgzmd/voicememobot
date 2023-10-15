package bot

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-audio/wav"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"voicesummary/config"
	"voicesummary/storage"
	stt "voicesummary/stt"
)

type BotProcessor struct {
	ctx          context.Context
	bot          *tgbotapi.BotAPI
	storage      storage.Storage
	speechToText stt.SpeechToText
	cfg          *config.Config
}

func NewBotProcessor(ctx context.Context, cfg *config.Config) (*BotProcessor, error) {
	bot, err := initializeBot(cfg.TelegramBotToken)
	if err != nil {
		return nil, err
	}

	stg, err := storage.NewRealStorage(ctx, cfg)
	if err != nil {
		return nil, err
	}

	speechToText, err := stt.NewGoogleSpeechToText(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &BotProcessor{
		ctx:          ctx,
		bot:          bot,
		storage:      stg,
		speechToText: speechToText,
		cfg:          cfg,
	}, nil
}

func (bp *BotProcessor) ProcessUpdatesLoop() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bp.bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Failed to start getting updates: %v", err)
	}

	for update := range updates {
		log.Printf("%+v", update)
		if update.Message != nil && (update.Message.Audio != nil || update.Message.Voice != nil) {
			bp.ProcessMessage(update)
		}
	}
}

func initializeBot(token string) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	bot.Debug = true
	if err == nil {
		log.Printf("Authorized on account %s", bot.Self.UserName)
	}
	return bot, err
}

// getAudioData takes a File object, downloads the associated audio file,
// and returns the file content as a byte slice or an error if one occurred.
func (bp *BotProcessor) getAudioData(audioFile *tgbotapi.File) ([]byte, error) {
	// Create the download link for the file
	lnk := audioFile.Link(bp.bot.Token)

	// Download the audio file.
	resp, err := http.Get(lnk)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio file: %v", err)
	}
	defer resp.Body.Close()

	// Read the audio file content.
	oggData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file content: %v", err)
	}

	return oggData, nil
}

func (bp *BotProcessor) getWavDuration(wavData []byte) (float64, error) {
	// Initialize a new WAV decoder
	decoder := wav.NewDecoder(bytes.NewReader(wavData))

	// Check if the WAV file is valid
	if !decoder.IsValidFile() {
		return 0, fmt.Errorf("invalid WAV file")
	}

	dur, err := decoder.Duration()
	if err != nil {
		return 0, fmt.Errorf("failed to get duration: %+v", err)
	} else {
		return dur.Seconds(), nil
	}
}

func (bp *BotProcessor) checkUser(update tgbotapi.Update) bool {
	for _, user := range bp.cfg.Usernames {
		if update.Message.From.UserName == user {
			return true
		}
	}

	return false
}

func (bp *BotProcessor) ProcessMessage(update tgbotapi.Update) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	if !bp.checkUser(update) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are not allowed to use this bot")
		bp.bot.Send(msg)

		log.Printf("User %s is not allowed to use the bot", update.Message.From.UserName)
		return
	}

	audioFileID := bp.getFileId(update)
	var file tgbotapi.File
	var err error
	file, err = bp.bot.GetFile(tgbotapi.FileConfig{FileID: audioFileID})

	if err != nil {
		log.Printf("Failed to get audio file path: %+v", err)
		return
	}
	audioData, err := bp.getAudioData(&file)
	if err != nil {
		log.Printf("Failed to get audio data: %v", err)
		return
	}

	wavData, err := bp.convertToWav(audioData)
	if err != nil {
		log.Printf("Failed to convert audio to WAV: %v", err)
		return
	}

	duration, err := bp.getWavDuration(wavData)
	if err != nil {
		log.Printf("Failed to get WAV duration: %v", err)
		return
	}

	var text string
	if duration > 60 {
		uri, err := bp.storage.StoreFile(bp.ctx, wavData)
		if err != nil {
			log.Printf("Failed to store audio file: %v", err)
			return
		}
		defer bp.storage.ClearFile(bp.ctx, uri)

		text, err = bp.speechToText.RecognizeSpeechFromFile(bp.ctx, uri)
		if err != nil {
			log.Printf("Failed to recognize speech: %v", err)
			return
		}
	} else {
		text, err = bp.speechToText.RecognizeSpeech(bp.ctx, wavData)
		if err != nil {
			log.Printf("Failed to recognize speech: %v", err)
			return
		}
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	bp.bot.Send(msg)
}

// convertToWav converts raw audio bytes to WAV format.
func (bp *BotProcessor) convertToWav(rawAudioBytes []byte) ([]byte, error) {
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
	err = bp.convertFileToWav(inp, wavFileName, err)
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
func (bp *BotProcessor) convertFileToWav(inp *os.File, wavFileName string, err error) error {
	cmd := exec.Command("ffmpeg", "-i", inp.Name(), "-acodec", "pcm_s16le", "-ac", "1", "-ar", "16000", "-y", wavFileName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	return err
}

// getFileId returns file ID of either voice memo or audio file sent by the user,
// depending on which one is present.
func (bp *BotProcessor) getFileId(update tgbotapi.Update) string {
	if update.Message.Voice != nil {
		return update.Message.Voice.FileID
	}
	return update.Message.Audio.FileID
}
