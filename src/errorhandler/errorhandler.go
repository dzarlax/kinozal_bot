package errorhandler

import (
	"kinozal-bot/errors"
	"kinozal-bot/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ErrorHandler struct {
	Bot *tgbotapi.BotAPI
}

func (eh *ErrorHandler) Handle(err error, chatID int64) {
	var userMessage string
	logFields := map[string]interface{}{
		"chat_id": chatID,
		"error":   err.Error(),
	}

	switch e := err.(type) {
	case *errors.BotError:
		userMessage = "Произошла ошибка в работе бота. Попробуйте позже."
		logFields["code"] = e.Code
		logFields["details"] = e.Details
	case *errors.KinozalError:
		userMessage = "Ошибка в работе с Kinozal. Попробуйте позже."
		logFields["details"] = e.Details
	case *errors.TransmissionError:
		userMessage = "Ошибка в работе с Transmission. Проверьте настройки."
		logFields["details"] = e.Details
	default:
		userMessage = "Неизвестная ошибка. Попробуйте позже."
	}

	logger.Error("Error handled", logFields)

	if eh.Bot != nil {
		msg := tgbotapi.NewMessage(chatID, userMessage)
		_, sendErr := eh.Bot.Send(msg)
		if sendErr != nil {
			logger.Warn("Failed to send error message to user", map[string]interface{}{
				"chat_id":     chatID,
				"error":       sendErr.Error(),
				"originalErr": err.Error(),
			})
		}
	}
}