package fileutils

import (
	"fmt"
	"os"
	"path/filepath"

	"kinozal-bot/errors"
	"kinozal-bot/logger"
)

// SaveTorrentFile saves the torrent file to disk
func SaveTorrentFile(kzID string, data []byte) (string, error) {
	// Определяем путь к файлу
	torrentDir := "torrents"
	torrentPath := filepath.Join(torrentDir, fmt.Sprintf("%s.torrent", kzID))

	// Проверяем, существует ли директория
	err := os.MkdirAll(torrentDir, os.ModePerm)
	if err != nil {
		return "", errors.NewBotError(
			"Failed to create directory for torrent files",
			"FILE_SYSTEM_ERROR",
			map[string]interface{}{
				"directory":      torrentDir,
				"original_error": err.Error(),
			},
		)
	}

	// Сохраняем файл
	err = os.WriteFile(torrentPath, data, 0644)
	if err != nil {
		return "", errors.NewBotError(
			"Failed to save torrent file",
			"FILE_SYSTEM_ERROR",
			map[string]interface{}{
				"kzID":           kzID,
				"original_error": err.Error(),
			},
		)
	}

	logger.Debug("Torrent file saved", map[string]interface{}{
		"path": torrentPath,
	})
	return torrentPath, nil
}

// CleanupTorrentFile removes the torrent file after use
func CleanupTorrentFile(torrentPath string) error {
	err := os.Remove(torrentPath)
	if err != nil {
		logger.Warn("Failed to cleanup torrent file", map[string]interface{}{
			"path":           torrentPath,
			"original_error": err.Error(),
		})
		return err
	}

	logger.Debug("Torrent file cleaned up", map[string]interface{}{
		"path": torrentPath,
	})
	return nil
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, errors.NewBotError(
			"Failed to get file size",
			"FILE_SYSTEM_ERROR",
			map[string]interface{}{
				"path":           filePath,
				"original_error": err.Error(),
			},
		)
	}
	return fileInfo.Size(), nil
}