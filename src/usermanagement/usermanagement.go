package usermanagement

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"kinozal-bot/config"
	"kinozal-bot/logger"
)

func handleAddUser(bot *tgbotapi.BotAPI, cfg *config.Config, chatID int64, args string) {
	userID, err := strconv.Atoi(args)
	if err != nil || userID <= 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Укажите корректный ID пользователя."))
		return
	}

	for _, id := range cfg.Bot.AllowedUsers {
		if id == userID {
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Пользователь %d уже добавлен.", userID)))
			return
		}
	}

	cfg.Bot.AllowedUsers = append(cfg.Bot.AllowedUsers, userID)
	if err := config.SaveUsersToFile(cfg); err != nil {
		logger.Error("Failed to save users", map[string]interface{}{
			"error": err.Error(),
		})
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка сохранения пользователей."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Пользователь %d добавлен в список разрешенных.", userID)))
}

func handleRemoveUser(bot *tgbotapi.BotAPI, cfg *config.Config, chatID int64, args string) {
	userID, err := strconv.Atoi(args)
	if err != nil || userID <= 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Укажите корректный ID пользователя."))
		return
	}

	found := false
	newAllowedUsers := []int{}
	for _, id := range cfg.Bot.AllowedUsers {
		if id != userID {
			newAllowedUsers = append(newAllowedUsers, id)
		} else {
			found = true
		}
	}

	if !found {
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Пользователь %d не найден в списке разрешенных.", userID)))
		return
	}

	cfg.Bot.AllowedUsers = newAllowedUsers
	if err := config.SaveUsersToFile(cfg); err != nil {
		logger.Error("Failed to save users", map[string]interface{}{
			"error": err.Error(),
		})
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка сохранения пользователей."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Пользователь %d удален из списка разрешенных.", userID)))
}

// HandleUserCommands обрабатывает команды для управления пользователями
// HandleUserCommands обрабатывает команды для управления пользователями
func HandleUserCommands(bot *tgbotapi.BotAPI, cfg *config.Config, chatID int64, command string, args string) {
	if chatID != int64(cfg.Bot.AdminID) {
		bot.Send(tgbotapi.NewMessage(chatID, "У вас нет прав для выполнения этой команды."))
		return
	}

	switch command {
	case "adduser":
		handleAddUser(bot, cfg, chatID, args)
	case "removeuser":
		handleRemoveUser(bot, cfg, chatID, args)
	case "listusers":
		handleListUsers(bot, cfg, chatID)
	default:
		bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда управления пользователями."))
	}
}

// handleListUsers отображает список разрешённых пользователей
func handleListUsers(bot *tgbotapi.BotAPI, cfg *config.Config, chatID int64) {
	if len(cfg.Bot.AllowedUsers) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Список разрешённых пользователей пуст."))
		return
	}

	var usersList []string
	for _, userID := range cfg.Bot.AllowedUsers {
		usersList = append(usersList, strconv.Itoa(userID))
	}
	message := "Разрешённые пользователи:\n" + strings.Join(usersList, "\n")

	bot.Send(tgbotapi.NewMessage(chatID, message))
}