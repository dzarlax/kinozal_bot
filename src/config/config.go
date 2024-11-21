package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	//"strings"

	"github.com/joho/godotenv"
)

// Config структура для хранения конфигурации
type Config struct {
	Telegram struct {
		Token   string
		Polling bool
	}
	Kinozal struct {
		Address   string
		Username  string
		Password  string
		Endpoints struct {
			Login   string
			Hash    string
			Details string
			Search  string
		}
	}
	Transmission struct {
		Host string
		Port int
		Auth struct {
			Username string
			Password string
		}
	}
	Folders struct {
		Torrents   string
		Films      string
		Series     string
		Audiobooks string
	}
	Bot struct {
		AdminID      int
		AllowedUsers []int // В памяти, но хранится в users.json
	}
}

const UsersFilePath = "/config/users.json"

// LoadConfig загружает конфигурацию из .env
func LoadConfig() (*Config, error) {
	// Попытка загрузки .env файла (для локальной среды)
	_ = godotenv.Load(".env") // Игнорируем ошибку, если файл не найден

	cfg := &Config{}

	// Чтение переменных окружения
	cfg.Telegram.Token = os.Getenv("TG_TOKEN")
	if cfg.Telegram.Token == "" {
		return nil, errors.New("TG_TOKEN is required")
	}
	cfg.Telegram.Polling = true

	cfg.Kinozal.Address = os.Getenv("KZ_ADDR")
	if cfg.Kinozal.Address == "" {
		cfg.Kinozal.Address = "kinozal.tv"
	}
	cfg.Kinozal.Username = os.Getenv("KZ_USER")
	cfg.Kinozal.Password = os.Getenv("KZ_PASS")
	if cfg.Kinozal.Username == "" || cfg.Kinozal.Password == "" {
		return nil, errors.New("KZ_USER and KZ_PASS are required")
	}
	cfg.Kinozal.Endpoints.Login = "/takelogin.php"
	cfg.Kinozal.Endpoints.Hash = "/get_srv_details.php"
	cfg.Kinozal.Endpoints.Details = "/details.php"
	cfg.Kinozal.Endpoints.Search = "/browse.php"

	cfg.Transmission.Host = os.Getenv("TRANS_ADDR")
	if cfg.Transmission.Host == "" {
		cfg.Transmission.Host = "localhost"
	}
	port, err := strconv.Atoi(os.Getenv("TRANS_PORT"))
	if err != nil {
		port = 9091
	}
	cfg.Transmission.Port = port
	cfg.Transmission.Auth.Username = os.Getenv("TRANS_USER")
	cfg.Transmission.Auth.Password = os.Getenv("TRANS_PASS")

	currentDir, _ := os.Getwd()
	cfg.Folders.Torrents = filepath.Join(currentDir, "torrents")
	cfg.Folders.Films = os.Getenv("FILMS_FOLDER")
	if cfg.Folders.Films == "" {
		cfg.Folders.Films = filepath.Join(currentDir, "downloads", "films")
	}
	cfg.Folders.Series = os.Getenv("SERIES_FOLDER")
	if cfg.Folders.Series == "" {
		cfg.Folders.Series = filepath.Join(currentDir, "downloads", "series")
	}
	cfg.Folders.Audiobooks = os.Getenv("AUDIOBOOKS_FOLDER")
	if cfg.Folders.Audiobooks == "" {
		cfg.Folders.Audiobooks = filepath.Join(currentDir, "downloads", "audiobooks")
	}

	adminID, err := strconv.Atoi(os.Getenv("BOT_ADMIN_ID"))
	if err != nil {
		return nil, errors.New("Invalid BOT_ADMIN_ID")
	}
	cfg.Bot.AdminID = adminID

	// Загружаем пользователей из файла
	if err := loadUsersFromFile(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadUsersFromFile загружает список разрешенных пользователей из файла
func loadUsersFromFile(cfg *Config) error {
	file, err := os.Open(UsersFilePath)
	if os.IsNotExist(err) {
		return nil // Если файла нет, просто возвращаем пустой список
	} else if err != nil {
		return err
	}
	defer file.Close()

	var users []int
	if err := json.NewDecoder(file).Decode(&users); err != nil {
		return err
	}

	cfg.Bot.AllowedUsers = users
	return nil
}

// SaveUsersToFile сохраняет список пользователей в файл
func SaveUsersToFile(cfg *Config) error {
	file, err := os.Create(UsersFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(cfg.Bot.AllowedUsers)
}
