package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"kinozal-bot/config"
	"kinozal-bot/errorhandler"
	"kinozal-bot/logger"
	"kinozal-bot/menu"
	"kinozal-bot/middleware"
	"kinozal-bot/torrent"
	"kinozal-bot/transmission"
	"kinozal-bot/usermanagement"
)

// Rate limiting for user requests
var (
	userLastRequest = make(map[int64]time.Time)
	userRequestMutex = sync.RWMutex{}
	requestCooldown = 10 * time.Second // 10 seconds cooldown between requests
)

// TelegramBotWrapper реализует интерфейс transmission.BotInterface
type TelegramBotWrapper struct {
	Bot *tgbotapi.BotAPI
}

func (tbw *TelegramBotWrapper) SendMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := tbw.Bot.Send(msg)
	return err
}

func (tbw *TelegramBotWrapper) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return tbw.Bot.Send(c)
}

func (tbw *TelegramBotWrapper) AnswerCallbackQuery(callbackConfig tgbotapi.CallbackConfig) (tgbotapi.APIResponse, error) {
	response, err := tbw.Bot.Request(callbackConfig)
	if err != nil {
		return tgbotapi.APIResponse{}, err
	}
	return *response, nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.Info("Starting bot", nil)

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		logger.Error("Failed to create Telegram bot instance", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to create bot: %v", err)
	}
	bot.Debug = false

	logger.Info("Bot authorized successfully", map[string]interface{}{
		"username": bot.Self.UserName,
	})

	eh := &errorhandler.ErrorHandler{Bot: bot}
	mw := &middleware.AccessMiddleware{Bot: bot, Cfg: cfg}

	// Устанавливаем команды для текущего чата
	err = menu.SetupBotCommands(bot, cfg, int64(cfg.Bot.AdminID))
	if err != nil {
		logger.Error("Failed to setup bot commands", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to setup bot commands: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range sigs {
			logger.Info("Shutting down bot", map[string]interface{}{
				"signal": sig.String(),
			})
			bot.StopReceivingUpdates()
			os.Exit(0)
		}
	}()

	wrappedBot := &TelegramBotWrapper{Bot: bot}

	for update := range updates {
		if update.Message != nil {
			logger.Info("Message received", map[string]interface{}{
				"chat_id": update.Message.Chat.ID,
				"text":    update.Message.Text,
			})
	
			if !mw.CheckAccess(int(update.Message.From.ID), update.Message.Chat.ID) {
				continue
			}
	
			switch update.Message.Command() {
			case "start":
				menu.HandleStart(bot, cfg, eh, update)
			case "help":
				menu.HandleHelp(bot, cfg, eh, update)
			case "find":
				handleFind(bot, cfg, eh, update)
			case "adduser", "removeuser", "listusers":
				usermanagement.HandleUserCommands(
					bot,
					cfg,
					update.Message.Chat.ID,
					update.Message.Command(),
					update.Message.CommandArguments(),
				)
			default:
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))
			}
		}
	
		if update.CallbackQuery != nil {
			handleCallback(wrappedBot, cfg, update.CallbackQuery)
		}
	}
}

func handleFind(bot *tgbotapi.BotAPI, cfg *config.Config, eh *errorhandler.ErrorHandler, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	query := update.Message.CommandArguments()

	if query == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, укажите поисковый запрос. Например: /find Матрица"))
		return
	}

	// Check rate limiting
	userRequestMutex.RLock()
	lastRequest, exists := userLastRequest[int64(userID)]
	userRequestMutex.RUnlock()
	
	if exists {
		timeSinceLastRequest := time.Since(lastRequest)
		if timeSinceLastRequest < requestCooldown {
			remainingTime := requestCooldown - timeSinceLastRequest
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("⏳ Пожалуйста, подождите %d секунд перед следующим поиском.", int(remainingTime.Seconds())+1)))
			return
		}
	}

	// Update last request time
	userRequestMutex.Lock()
	userLastRequest[int64(userID)] = time.Now()
	userRequestMutex.Unlock()

	// Notify user that search is starting
	searchingMsg := tgbotapi.NewMessage(chatID, "🔍 Выполняется поиск, пожалуйста подождите...")
	sentMsg, err := bot.Send(searchingMsg)
	if err != nil {
		logger.Error("Failed to send searching message", map[string]interface{}{
			"error": err.Error(),
		})
	}

	client, _, err := torrent.LoginKinozal(cfg)
	if err != nil {
		// Delete searching message and send error
		if sentMsg.MessageID != 0 {
			deleteMsg := tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID)
			bot.Send(deleteMsg)
		}
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка при входе на Kinozal. Проверьте подключение к интернету и попробуйте позже."))
		eh.Handle(err, chatID)
		return
	}

	results, err := torrent.SearchTorrents(cfg, client, query)
	if err != nil {
		// Delete searching message and send error
		if sentMsg.MessageID != 0 {
			deleteMsg := tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID)
			bot.Send(deleteMsg)
		}
		
		// Check if it's a rate limiting issue (400 error)
		if strings.Contains(err.Error(), "400 Bad Request") {
			bot.Send(tgbotapi.NewMessage(chatID, "⏳ Слишком много запросов к Kinozal. Подождите немного и попробуйте снова."))
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка поиска на Kinozal. Попробуйте изменить поисковый запрос или повторить попытку позже."))
		}
		eh.Handle(err, chatID)
		return
	}

	// Delete searching message
	if sentMsg.MessageID != 0 {
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID)
		bot.Send(deleteMsg)
	}

	if len(results) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ничего не найдено по вашему запросу. Попробуйте изменить поисковые слова."))
		return
	}

	sendSearchResults(bot, chatID, results)
}

func sendSearchResults(bot *tgbotapi.BotAPI, chatID int64, results []torrent.SearchResult) {
	topResults := results
	if len(results) > 7 {
		topResults = results[:7]
	}

	var keyboardRows [][]tgbotapi.InlineKeyboardButton

	messageText := "🔍 Найденные результаты:\n\n"
	for _, result := range topResults {
		messageText += fmt.Sprintf("🎬 %s\nSeeders: %d | Size: %s\n\n", result.Title, result.Seeders, result.Size)
		button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("⬇ Скачать: %s", result.Title), fmt.Sprintf("startdownload_%s", result.ID))
		keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(button))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		logger.Error("Failed to send search results", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

func handleCallback(bot transmission.BotInterface, cfg *config.Config, callback *tgbotapi.CallbackQuery) {
	data := callback.Data
	chatID := callback.Message.Chat.ID

	logger.Debug("Received callback data", map[string]interface{}{
		"data": data,
	})

	if strings.HasPrefix(data, "startdownload_") {
		kzID := strings.TrimPrefix(data, "startdownload_")
		logger.Debug("Download button pressed", map[string]interface{}{
			"kzID": kzID,
		})
	
		if kzID == "" {
			logger.Error("Invalid callback data: missing kzID", nil)
			bot.SendMessage(chatID, "Ошибка: не указан ID раздачи.")
			return
		}
	
		kzName := fmt.Sprintf("Раздача-%s", kzID)
		logger.Debug("Parsed callback data", map[string]interface{}{
			"kzID":   kzID,
			"kzName": kzName,
		})
	
		// Создаем HTTP клиент
		httpClient := &http.Client{}
		torrentPath, err := torrent.DownloadTorrent(cfg, httpClient, kzID)
		if err != nil {
			logger.Error("Failed to download torrent", map[string]interface{}{
				"error":      err.Error(),
				"torrent_id": kzID,
			})
			bot.SendMessage(chatID, fmt.Sprintf("Ошибка загрузки торрента: %s", err.Error()))
			return
		}
	
		logger.Info("Torrent downloaded successfully", map[string]interface{}{
			"torrent_path": torrentPath,
			"kzID":         kzID,
		})
	
		// Формирование списка папок для выбора
		var keyboardRows [][]tgbotapi.InlineKeyboardButton
		for _, folder := range []struct {
			Name string
			Path string
		}{
			{"Фильмы", cfg.Folders.Films},
			{"Сериалы", cfg.Folders.Series},
			{"Аудиокниги", cfg.Folders.Audiobooks},
		} {
			if folder.Path == "" {
				logger.Warn("Skipping empty folder path", map[string]interface{}{
					"folder_name": folder.Name,
				})
				continue
			}
			button := tgbotapi.NewInlineKeyboardButtonData(folder.Name, fmt.Sprintf("selectfolder_%s_%s_%s", kzID, kzName, folder.Path))
			keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(button))
		}
	
		if len(keyboardRows) == 0 {
			bot.SendMessage(chatID, "Нет доступных папок для загрузки.")
			return
		}
	
		msg := tgbotapi.NewMessage(chatID, "Выберите папку для загрузки:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
		if _, err := bot.Send(msg); err != nil {
			logger.Error("Failed to send folder selection buttons", map[string]interface{}{
				"error": err.Error(),
			})
			bot.SendMessage(chatID, "Произошла ошибка при отображении списка папок.")
		}
	}
	if strings.HasPrefix(data, "selectfolder_") {
		logger.Debug("Folder selection detected", map[string]interface{}{
			"data": data,
		})
	
		parts := strings.Split(data, "_")
		if len(parts) < 4 {
			logger.Error("Invalid callback data for folder selection", map[string]interface{}{
				"data": data,
			})
			bot.SendMessage(chatID, "Ошибка: Неверные данные для выбора папки.")
			return
		}
	
		kzID := parts[1]
		kzName := parts[2]
		folderPath := strings.Join(parts[3:], "_")
	
		logger.Debug("Parsed folder selection data", map[string]interface{}{
			"kzID":       kzID,
			"kzName":     kzName,
			"folderPath": folderPath,
		})
	
		torrentPath := fmt.Sprintf("torrents/%s.torrent", kzID)
		if err := transmission.AddToTransmission(torrentPath, folderPath, kzName, chatID, bot); err != nil {
			logger.Error("Failed to add torrent to Transmission", map[string]interface{}{
				"error":        err.Error(),
				"torrent_path": torrentPath,
				"folder_path":  folderPath,
			})
			bot.SendMessage(chatID, fmt.Sprintf("Ошибка добавления в Transmission: %s", err.Error()))
			return
		}
	
		bot.SendMessage(chatID, fmt.Sprintf("Торрент %s добавлен в папку %s.", kzName, folderPath))
	}
}