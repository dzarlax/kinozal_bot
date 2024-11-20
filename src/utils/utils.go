package utils

import (
	"bytes"
	"strings"
	"unicode/utf8"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"kinozal-bot/torrent"
)

func toUTF8(input string) string {
	if utf8.ValidString(input) {
		return input
	}
	// Преобразуем некорректные строки
	buf := bytes.NewBuffer(nil)
	for _, r := range input {
		buf.WriteRune(r)
	}
	return buf.String()
}

func SendSearchResults(bot *tgbotapi.BotAPI, chatID int64, results []torrent.SearchResult) {
	const maxMessageLength = 4000 // Telegram limit is 4096
	var messageBuilder strings.Builder

	for i, result := range results {
		line := fmt.Sprintf("%d. %s\n", i+1, toUTF8(result.Title))
		if messageBuilder.Len()+len(line) > maxMessageLength {
			// Отправляем текущую часть
			msg := tgbotapi.NewMessage(chatID, toUTF8(messageBuilder.String()))
			bot.Send(msg)
			messageBuilder.Reset()
		}
		messageBuilder.WriteString(line)
	}

	// Отправляем оставшиеся результаты
	if messageBuilder.Len() > 0 {
		msg := tgbotapi.NewMessage(chatID, toUTF8(messageBuilder.String()))
		bot.Send(msg)
	}
}