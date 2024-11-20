package menu

import (
	"fmt"
	"strings"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"kinozal-bot/config"
	"kinozal-bot/errorhandler"
	"kinozal-bot/logger"
	"kinozal-bot/usermanagement"
)

func SetupBotCommands(bot *tgbotapi.BotAPI, cfg *config.Config, chatID int64) error {
	// Базовые команды, которые доступны всем пользователям
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Запустить бота и получить информацию"},
		{Command: "help", Description: "Показать справку по использованию"},
		{Command: "find", Description: "Найти торрент (например: /find Матрица)"},
	}

	// Если пользователь — администратор, добавляем команды управления пользователями
	if chatID == int64(cfg.Bot.AdminID) {
		commands = append(commands,
			tgbotapi.BotCommand{Command: "adduser", Description: "Добавить пользователя в список разрешенных"},
			tgbotapi.BotCommand{Command: "removeuser", Description: "Удалить пользователя из списка разрешенных"},
			tgbotapi.BotCommand{Command: "listusers", Description: "Показать список разрешенных пользователей"},
		)
	}

	// Устанавливаем команды для текущего пользователя
	cmdConfig := tgbotapi.NewSetMyCommands(commands...)
	_, err := bot.Request(cmdConfig)
	if err != nil {
		logger.Error("Failed to set bot commands", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	logger.Info("Bot commands successfully set up", nil)
	return nil
}

func HandleStart(bot *tgbotapi.BotAPI, cfg *config.Config, eh *errorhandler.ErrorHandler, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	username := update.Message.From.UserName

	// Сообщение с приветствием
	message := fmt.Sprintf(`Привет, %s\! 👋
🤖 Я бот для работы с торрентами Kinozal\.

*Мои возможности:*
🔍 Найдите раздачи: /find \<поисковый запрос\>
🆘 Получите справку: /help

`, escapeMarkdownV2(username))

	// Если пользователь администратор, добавляем инструкции
	if chatID == int64(cfg.Bot.AdminID) {
		message += `👤 *Администрирование:*
▫️ /adduser \<ID\> \- добавить пользователя
▫️ /removeuser \<ID\> \- удалить пользователя
▫️ /listusers \- показать список пользователей\.
`
	}

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "MarkdownV2"

	if _, err := bot.Send(msg); err != nil {
		logger.Error("Failed to send /start response", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// Функция для экранирования специальных символов в MarkdownV2
func escapeMarkdownV2(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}

func HandleHelp(bot *tgbotapi.BotAPI, cfg *config.Config, eh *errorhandler.ErrorHandler, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	helpMessage := `📖 Справка по командам:

🔍 *Поиск торрентов*
/find - поиск торрентов по названию (например: /find Матрица)
`

	// Добавляем админские команды в справку, если пользователь — администратор
	if chatID == int64(cfg.Bot.AdminID) {
		helpMessage += `👤 *Администрирование:*
/adduser <ID> - добавить пользователя
/removeuser <ID> - удалить пользователя
/listusers - показать список пользователей.
`
	}

	msg := tgbotapi.NewMessage(chatID, helpMessage)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func HandleUserCommands(bot *tgbotapi.BotAPI, cfg *config.Config, eh *errorhandler.ErrorHandler, update tgbotapi.Update) {
	switch update.Message.Command() {
	case "adduser", "removeuser", "listusers":
		usermanagement.HandleUserCommands(
			bot,
			cfg,
			update.Message.Chat.ID,
			update.Message.Command(),
			update.Message.CommandArguments(),
		)
	}
}