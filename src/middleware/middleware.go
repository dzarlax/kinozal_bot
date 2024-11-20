package middleware

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"kinozal-bot/config"
	"kinozal-bot/logger"
)

type AccessMiddleware struct {
	Bot *tgbotapi.BotAPI
	Cfg *config.Config
}

func (am *AccessMiddleware) CheckAccess(userID int, chatID int64) bool {
	isAllowed := am.isAllowedUser(userID)

	if !isAllowed {
		logger.Warn("Unauthorized access attempt", map[string]interface{}{
			"user_id": userID,
			"chat_id": chatID,
		})

		msg := tgbotapi.NewMessage(chatID, "У вас нет доступа к этому боту. Пожалуйста, обратитесь к администратору.")
		_, err := am.Bot.Send(msg)
		if err != nil {
			logger.Error("Failed to send access denial message", map[string]interface{}{
				"error": err.Error(),
			})
		}

		if am.Cfg.Bot.AdminID != 0 {
			adminMsg := fmt.Sprintf("⚠️ Попытка несанкционированного доступа:\nUser ID: %d\nChat ID: %d", userID, chatID)
			adminMessage := tgbotapi.NewMessage(int64(am.Cfg.Bot.AdminID), adminMsg)
			_, err := am.Bot.Send(adminMessage)
			if err != nil {
				logger.Warn("Failed to notify admin about unauthorized access", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}

	return isAllowed
}

func (am *AccessMiddleware) isAllowedUser(userID int) bool {
	if userID == am.Cfg.Bot.AdminID {
		return true
	}

	for _, allowedID := range am.Cfg.Bot.AllowedUsers {
		if allowedID == userID {
			return true
		}
	}

	return false
}