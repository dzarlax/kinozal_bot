package torrent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	"kinozal-bot/config"
	"kinozal-bot/errors"
	"kinozal-bot/fileutils"
	"kinozal-bot/logger"

	"github.com/PuerkitoBio/goquery"
)

type SearchResult struct {
	Title   string
	ID      string
	Seeders int
	Size    string
}

// Создаем HTTP-клиент с тайм-аутами
func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

// SaveCookies сохраняет куки в файл
func SaveCookies(cookies []*http.Cookie, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Failed to create cookie file: %w", err)
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(cookies)
	if err != nil {
		return fmt.Errorf("Failed to encode cookies to JSON: %w", err)
	}

	return nil
}

// LoadCookies загружает куки из файла
func LoadCookies(filePath string) ([]*http.Cookie, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open cookie file: %w", err)
	}
	defer file.Close()

	var cookies []*http.Cookie
	err = json.NewDecoder(file).Decode(&cookies)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode cookies from JSON: %w", err)
	}

	return cookies, nil
}

func LoginKinozal(cfg *config.Config) (*http.Client, []*http.Cookie, error) {
	// Создаем CookieJar для хранения кук
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create cookie jar: %w", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	// Шаг 1: Тестирование соединения с главной страницей
	mainPageURL := fmt.Sprintf("https://%s/", cfg.Kinozal.Address)
	logger.Debug("Testing connection to main page", map[string]interface{}{
		"url": mainPageURL,
	})

	resp, err := client.Get(mainPageURL)
	if err != nil {
		return nil, nil, errors.NewKinozalError("Failed to connect to main page", map[string]interface{}{"error": err.Error()})
	}
	defer resp.Body.Close()

	logger.Debug("Main page response", map[string]interface{}{
		"status":  resp.StatusCode,
		"cookies": jar.Cookies(resp.Request.URL),
	})

	// Шаг 2: Отправка запроса логина
	loginURL := fmt.Sprintf("https://%s%s", cfg.Kinozal.Address, cfg.Kinozal.Endpoints.Login)
	loginData := fmt.Sprintf("username=%s&password=%s&returnto=/&before=//&auth_submit_login=submit",
		cfg.Kinozal.Username, cfg.Kinozal.Password)

	req, err := http.NewRequest("POST", loginURL, bytes.NewBufferString(loginData))
	if err != nil {
		return nil, nil, errors.NewKinozalError("Failed to create login request", map[string]interface{}{"error": err.Error()})
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", mainPageURL)
	req.Header.Set("Origin", mainPageURL)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	resp, err = client.Do(req)
	if err != nil {
		return nil, nil, errors.NewKinozalError("Failed to execute login request", map[string]interface{}{"error": err.Error()})
	}
	defer resp.Body.Close()

	// Чтение тела ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.NewKinozalError("Failed to read login response", map[string]interface{}{"error": err.Error()})
	}

	logger.Debug("Login response body", map[string]interface{}{
		"body": string(body),
	})

	// Шаг 3: Проверка кук
	cookies := jar.Cookies(resp.Request.URL)
	hasUID := false
	hasPass := false

	for _, cookie := range cookies {
		if cookie.Name == "uid" {
			hasUID = true
		}
		if cookie.Name == "pass" {
			hasPass = true
		}
	}

	logger.Debug("Cookie check after login", map[string]interface{}{
		"has_uid":  hasUID,
		"has_pass": hasPass,
	})

	if !hasUID || !hasPass {
		return nil, nil, errors.NewKinozalError("Login failed - missing required cookies", map[string]interface{}{
			"cookies": cookies,
		})
	}

	// Сохраняем куки в файл
	const cookieFilePath = "kinozal_cookies.json"
	err = SaveCookies(cookies, cookieFilePath)
	if err != nil {
		logger.Warn("Failed to save cookies after login", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		logger.Info("Cookies saved successfully after login", map[string]interface{}{
			"file": cookieFilePath,
		})
	}

	logger.Info("Successfully logged in to Kinozal", nil)

	// Возвращаем клиент и куки
	return client, cookies, nil
}

func SearchTorrents(cfg *config.Config, client *http.Client, query string) ([]SearchResult, error) {
	// Add delay to prevent rate limiting
	time.Sleep(1 * time.Second)
	
	// Properly URL encode the query to handle spaces and special characters
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("https://%s%s?s=%s", cfg.Kinozal.Address, cfg.Kinozal.Endpoints.Search, encodedQuery)

	logger.Debug("Searching torrents", map[string]interface{}{
		"url":           searchURL,
		"query":         query,
		"encoded_query": encodedQuery,
	})

	// Create a new request to set headers
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, errors.NewKinozalError("Failed to create search request", map[string]interface{}{"error": err.Error()})
	}

	// Set headers to mimic browser behavior
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	// Remove Accept-Encoding header to prevent gzip compression issues
	// req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/", cfg.Kinozal.Address))

	// Retry mechanism for handling temporary failures
	maxRetries := 3
	var resp *http.Response
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Debug("Search attempt", map[string]interface{}{
			"attempt": attempt,
			"max_retries": maxRetries,
		})
		
		// Execute the request
		resp, err = client.Do(req)
		if err != nil {
			if attempt == maxRetries {
				return nil, errors.NewKinozalError("Failed to execute search request after retries", map[string]interface{}{
					"error": err.Error(),
					"attempts": attempt,
				})
			}
			logger.Warn("Search request failed, retrying", map[string]interface{}{
				"error": err.Error(),
				"attempt": attempt,
			})
			time.Sleep(time.Duration(attempt) * 2 * time.Second) // Progressive delay
			continue
		}
		
		// Check response status
		if resp.StatusCode == http.StatusOK {
			break // Success, proceed with processing
		}
		
		resp.Body.Close() // Close body before retry
		
		if resp.StatusCode == 400 {
			if attempt == maxRetries {
				return nil, errors.NewKinozalError("Search request failed with 400 Bad Request", map[string]interface{}{
					"status": resp.Status,
					"attempts": attempt,
					"query": query,
					"encoded_query": encodedQuery,
				})
			}
			logger.Warn("Received 400 Bad Request, retrying with delay", map[string]interface{}{
				"attempt": attempt,
				"status": resp.Status,
			})
			time.Sleep(time.Duration(attempt) * 3 * time.Second) // Longer delay for 400 errors
			continue
		}
		
		if attempt == maxRetries {
			return nil, errors.NewKinozalError("Search request failed", map[string]interface{}{
				"status": resp.Status,
				"attempts": attempt,
			})
		}
		
		logger.Warn("Non-OK status, retrying", map[string]interface{}{
			"status": resp.Status,
			"attempt": attempt,
		})
		time.Sleep(time.Duration(attempt) * 2 * time.Second)
	}
	
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewKinozalError("Failed to read search response", map[string]interface{}{"error": err.Error()})
	}

	// Check if the response contains a login form, which would indicate authentication issues
	if strings.Contains(string(body), "name=\"username\"") && strings.Contains(string(body), "name=\"password\"") {
		logger.Warn("Received login page instead of search results, retrying login", nil)

		// Re-login and try again
		newClient, _, err := LoginKinozal(cfg)
		if err != nil {
			return nil, errors.NewKinozalError("Failed to re-login for search", map[string]interface{}{"error": err.Error()})
		}

		// Recursive call with the new client (only once to avoid infinite recursion)
		return SearchTorrents(cfg, newClient, query)
	}

	decodedBody, err := decodeWindows1251(string(body))
	if err != nil {
		return nil, errors.NewKinozalError("Failed to decode response from Windows-1251 to UTF-8", map[string]interface{}{"error": err.Error()})
	}

	results := ParseSearchResults(decodedBody)

	logger.Debug("Search results", map[string]interface{}{
		"count": len(results),
	})

	return results, nil
}

func DownloadTorrent(cfg *config.Config, client *http.Client, torrentID string) (string, error) {
	const cookieFilePath = "kinozal_cookies.json"

	logger.Debug("Starting torrent download", map[string]interface{}{
		"torrent_id": torrentID,
	})

	// Загружаем куки из файла, если файл существует
	var cookies []*http.Cookie
	if _, err := os.Stat(cookieFilePath); err == nil {
		cookies, err = LoadCookies(cookieFilePath)
		if err != nil {
			logger.Warn("Failed to load cookies", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logger.Info("Loaded cookies from file", map[string]interface{}{
				"file": cookieFilePath,
			})
		}
	}

	// Формируем URL для скачивания
	downloadURL := fmt.Sprintf("https://dl.%s/download.php?id=%s", cfg.Kinozal.Address, torrentID)
	logger.Debug("Download URL", map[string]interface{}{
		"url": downloadURL,
	})

retryDownload:
	// Создаём запрос
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to create download request: %w", err)
	}

	// Добавляем куки к запросу
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// Устанавливаем заголовки
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/details.php?id=%s", cfg.Kinozal.Address, torrentID))
	req.Header.Set("Sec-CH-UA", `"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to execute download request: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		logger.Warn("Non-OK status code from download request", map[string]interface{}{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
		})
	}

	// Проверяем содержимое ответа
	contentType := resp.Header.Get("Content-Type")
	logger.Debug("Download response content type", map[string]interface{}{
		"content_type": contentType,
	})

	if contentType != "application/x-bittorrent" {
		htmlBody, _ := ioutil.ReadAll(resp.Body)
		bodyPreview := string(htmlBody)
		if len(bodyPreview) > 200 {
			bodyPreview = bodyPreview[:200] + "..." // Truncate for logging
		}

		logger.Debug("Received non-torrent response", map[string]interface{}{
			"body_preview": bodyPreview,
		})

		// Если это HTML, проверяем авторизацию
		if strings.Contains(string(htmlBody), "name=\"username\"") ||
			strings.Contains(string(htmlBody), "name=\"password\"") ||
			strings.Contains(string(htmlBody), "login") ||
			strings.Contains(string(htmlBody), "вход") {
			logger.Warn("Authorization required, retrying login", map[string]interface{}{
				"torrent_id": torrentID,
			})

			client, cookies, err = LoginKinozal(cfg)
			if err != nil {
				return "", fmt.Errorf("Failed to re-login: %w", err)
			}

			// Повторяем запрос после повторного логина
			goto retryDownload
		}

		logger.Error("Unexpected HTML response", map[string]interface{}{
			"body": string(htmlBody),
		})
		return "", fmt.Errorf("Received HTML instead of torrent file")
	}

	// Читаем данные файла
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read torrent file data: %w", err)
	}

	// Проверяем размер данных
	if len(data) == 0 {
		return "", fmt.Errorf("Received empty torrent file")
	}

	logger.Debug("Torrent file data received", map[string]interface{}{
		"size_bytes": len(data),
	})

	// Сохраняем файл
	torrentPath, err := fileutils.SaveTorrentFile(torrentID, data)
	if err != nil {
		return "", fmt.Errorf("Failed to save torrent file: %w", err)
	}

	// Сохраняем актуальные куки после успешного запроса
	err = SaveCookies(resp.Cookies(), cookieFilePath)
	if err != nil {
		logger.Warn("Failed to save updated cookies", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		logger.Info("Updated cookies saved successfully", map[string]interface{}{
			"file": cookieFilePath,
		})
	}

	logger.Info("Torrent downloaded successfully", map[string]interface{}{
		"torrent_id": torrentID,
		"path":       torrentPath,
	})
	return torrentPath, nil
}

func ParseSearchResults(html string) []SearchResult {
	var results []SearchResult

	// Check if the HTML contains a login form, which would indicate we're not getting search results
	if strings.Contains(html, "name=\"username\"") && strings.Contains(html, "name=\"password\"") {
		logger.Warn("HTML contains login form instead of search results", nil)
		return results
	}

	// Check if the HTML contains an error message
	if strings.Contains(html, "По Вашему запросу ничего не найдено") {
		logger.Info("No results found for the search query", nil)
		return results
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		logger.Error("Failed to parse HTML", map[string]interface{}{
			"error": err.Error(),
		})
		return results
	}

	// Log detailed HTML structure for debugging
	logger.Debug("HTML structure analysis", map[string]interface{}{
		"html_length": len(html),
		"title":       doc.Find("title").Text(),
	})

	// Look for different possible table structures
	tableSelectors := []string{
		"tr.bg",           // Original selector
		"tr[class*='bg']", // Any class containing 'bg'
		"tr.row",          // Alternative class name
		"tr[style]",       // Rows with inline styles
		"table tr",        // Any table rows
		".torrent-row",    // Common torrent site pattern
		".torrent",        // Another common pattern
	}

	for _, selector := range tableSelectors {
		rowCount := doc.Find(selector).Length()
		logger.Debug("Testing selector", map[string]interface{}{
			"selector": selector,
			"count":    rowCount,
		})
		
		if rowCount > 0 {
			// Log first few elements for analysis
			doc.Find(selector).Each(func(i int, s *goquery.Selection) {
				if i < 3 { // Only log first 3 for brevity
					html, _ := s.Html()
					logger.Debug("Row HTML sample", map[string]interface{}{
						"index":    i,
						"selector": selector,
						"html":     html[:min(200, len(html))],
					})
				}
			})
		}
	}

	// Log the number of rows found with original selector
	rowCount := doc.Find("tr.bg").Length()
	logger.Debug("Found rows in search results", map[string]interface{}{
		"count": rowCount,
	})

	if rowCount == 0 {
		// If no rows found, log more detailed HTML sample
		htmlPreview := html
		if len(htmlPreview) > 1000 {
			htmlPreview = htmlPreview[:1000] + "..."
		}
		logger.Debug("No rows found in HTML, detailed preview:", map[string]interface{}{
			"html_preview": htmlPreview,
		})
		
		// Try to find any table structures
		doc.Find("table").Each(func(i int, table *goquery.Selection) {
			tableHtml, _ := table.Html()
			if len(tableHtml) > 500 {
				tableHtml = tableHtml[:500] + "..."
			}
			logger.Debug("Found table", map[string]interface{}{
				"table_index": i,
				"table_html":  tableHtml,
			})
		})
		
		return results
	}

	doc.Find("tr.bg").Each(func(index int, row *goquery.Selection) {
		// Название и ссылка
		title := strings.TrimSpace(row.Find("td.nam a").Text())
		href, exists := row.Find("td.nam a").Attr("href")
		if !exists {
			logger.Debug("No href found for row", map[string]interface{}{
				"index": index,
			})
			return
		}

		// Сидеры
		seedersText := strings.TrimSpace(row.Find("td.sl_s").Text())
		seeders := 0
		_, err := fmt.Sscanf(seedersText, "%d", &seeders)
		if err != nil {
			logger.Debug("Failed to parse seeders", map[string]interface{}{
				"text":  seedersText,
				"error": err.Error(),
			})
		}

		// Размер
		size := strings.TrimSpace(row.Find("td.s").Eq(1).Text()) // Второй <td> с классом "s" содержит размер
		if size == "" {
			logger.Debug("No size found for row", map[string]interface{}{
				"index": index,
			})
		}

		// ID торрента
		torrentID := extractIDFromHref(href)
		if torrentID == "" {
			logger.Debug("Failed to extract torrent ID from href", map[string]interface{}{
				"href": href,
			})
			return
		}

		result := SearchResult{
			Title:   title,
			ID:      torrentID,
			Seeders: seeders,
			Size:    size,
		}

		logger.Debug("Parsed search result", map[string]interface{}{
			"title":   title,
			"id":      torrentID,
			"seeders": seeders,
			"size":    size,
		})

		results = append(results, result)
	})

	// Sort results by seeders (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Seeders > results[j].Seeders
	})

	logger.Info("Parsed search results", map[string]interface{}{
		"total_results": len(results),
	})

	return results
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractIDFromHref extracts the torrent ID from the href attribute
func extractIDFromHref(href string) string {
	parts := strings.Split(href, "=")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// decodeWindows1251 converts Windows-1251 encoded string to UTF-8
func decodeWindows1251(input string) (string, error) {
	reader := strings.NewReader(input)
	decoder := charmap.Windows1251.NewDecoder()
	result, err := ioutil.ReadAll(transform.NewReader(reader, decoder))
	return string(result), err
}
