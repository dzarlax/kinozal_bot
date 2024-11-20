package transmission

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hekmon/transmissionrpc"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"kinozal-bot/fileutils"
	"kinozal-bot/logger"
	"kinozal-bot/errors"
	"kinozal-bot/config"
)

// BotInterface определяет необходимые методы для взаимодействия с Telegram Bot API
type BotInterface interface {
	SendMessage(chatID int64, message string) error
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	AnswerCallbackQuery(callbackConfig tgbotapi.CallbackConfig) (tgbotapi.APIResponse, error)
}

// fileToBase64 конвертирует файл в строку Base64
func fileToBase64(file *os.File) (string, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(content), nil
}

// AddToTransmission добавляет торрент в Transmission
func AddToTransmission(torrentPath, downloadPath, kzName string, chatID int64, bot BotInterface) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к Transmission RPC
	client, err := transmissionrpc.New(cfg.Transmission.Host, cfg.Transmission.Auth.Username, cfg.Transmission.Auth.Password, &transmissionrpc.AdvancedConfig{
		HTTPS: false,
		Port:  uint16(cfg.Transmission.Port), // Исправлено
	})
	if err != nil {
		return errors.NewTransmissionError("Failed to connect to Transmission RPC", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Открываем файл торрента
	torrentFile, err := os.Open(torrentPath)
	if err != nil {
		return errors.NewTransmissionError("Failed to open torrent file", map[string]interface{}{
			"torrent_path": torrentPath,
			"error":        err.Error(),
		})
	}
	defer torrentFile.Close()

	// Конвертируем торрент-файл в Base64
	metaInfo, err := fileToBase64(torrentFile)
	if err != nil {
		return errors.NewTransmissionError("Failed to encode torrent file to Base64", map[string]interface{}{
			"torrent_path": torrentPath,
			"error":        err.Error(),
		})
	}

	// Добавляем торрент в Transmission
	_, err = client.TorrentAdd(&transmissionrpc.TorrentAddPayload{
		DownloadDir: &downloadPath,
		MetaInfo:    &metaInfo, // Исправлено
	})
	if err != nil {
		return errors.NewTransmissionError("Failed to add torrent to Transmission", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Удаляем торрент-файл после добавления
	err = fileutils.CleanupTorrentFile(torrentPath)
	if err != nil {
		logger.Error("Failed to cleanup torrent file", map[string]interface{}{
			"torrent_path": torrentPath,
			"error":        err.Error(),
		})
	}

	// Отправляем сообщение пользователю
	message := fmt.Sprintf("Торрент %s добавлен в Transmission и будет загружен в папку \"%s\".", kzName, downloadPath)
	err = bot.SendMessage(chatID, message)
	if err != nil {
		logger.Error("Failed to send message to Telegram", map[string]interface{}{
			"chat_id": chatID,
			"error":   err.Error(),
		})
	}

	logger.Info("Torrent added to Transmission successfully", map[string]interface{}{
		"torrent_name": kzName,
		"download_dir": downloadPath,
	})
	return nil
}