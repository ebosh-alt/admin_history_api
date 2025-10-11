package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"admin_history/config"

	"go.uber.org/zap"
)

type Client struct {
	apiURL     string
	httpClient *http.Client
	log        *zap.Logger
	disabled   bool
}

type Option func(*Client)

func WithHTTPClient(cli *http.Client) Option {
	return func(c *Client) {
		if cli != nil {
			c.httpClient = cli
		}
	}
}

func NewClient(cfg *config.Config, log *zap.Logger, opts ...Option) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil config")
	}
	if log == nil {
		return nil, fmt.Errorf("nil logger")
	}

	token := strings.TrimSpace(cfg.Telegram.Token)

	baseURL := strings.TrimSpace(cfg.Telegram.APIBaseURL)
	if baseURL == "" {
		baseURL = "https://api.telegram.org"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	timeout := cfg.Telegram.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &Client{
		httpClient: &http.Client{Timeout: timeout},
		log:        log.Named("telegram"),
	}

	if token == "" {
		client.disabled = true
		client.log.Warn("telegram token is empty; notifications disabled")
	} else {
		client.apiURL = fmt.Sprintf("%s/bot%s", baseURL, token)
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

func (c *Client) SendDemoMedia(ctx context.Context, chatID int64, questionnaireID int64, photoPaths []string, videoPath string, caption string) error {
	if c == nil || c.disabled {
		return nil
	}

	for _, p := range photoPaths {
		path := strings.TrimSpace(p)
		if path == "" {
			continue
		}
		if err := c.sendPhoto(ctx, chatID, path); err != nil {
			return err
		}
	}

	videoPath = strings.TrimSpace(videoPath)
	if videoPath == "" {
		return nil
	}

	markup := inlineKeyboardMarkup{
		InlineKeyboard: [][]inlineKeyboardButton{
			{
				{Text: "💳 Оплатить", CallbackData: fmt.Sprintf("payment:yes:%d", questionnaireID)},
			},
			{
				{Text: "❌ Отказаться", CallbackData: fmt.Sprintf("payment:no:%d", questionnaireID)},
			},
			{
				{Text: "🔗 Больше примеров", URL: "https://instagram.com/istoriym_bot"},
			},
		},
	}

	return c.sendVideo(ctx, chatID, videoPath, caption, markup)
}

func (c *Client) sendPhoto(ctx context.Context, chatID int64, filePath string) error {
	if c.disabled {
		return nil
	}

	fields := map[string]string{
		"chat_id": strconv.FormatInt(chatID, 10),
	}
	return c.sendMultipart(ctx, "sendPhoto", fields, []fileField{
		{Field: "photo", Path: filePath},
	})
}

func (c *Client) sendVideo(ctx context.Context, chatID int64, filePath string, caption string, markup inlineKeyboardMarkup) error {
	if c.disabled {
		return nil
	}

	fields := map[string]string{
		"chat_id": strconv.FormatInt(chatID, 10),
	}

	if caption != "" {
		fields["caption"] = caption
	}

	if kb, err := json.Marshal(markup); err == nil {
		fields["reply_markup"] = string(kb)
	}

	return c.sendMultipart(ctx, "sendVideo", fields, []fileField{
		{Field: "video", Path: filePath},
	})
}

type fileField struct {
	Field string
	Path  string
}

type tgResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func (c *Client) sendMultipart(ctx context.Context, method string, fields map[string]string, files []fileField) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			writer.Close()
			return fmt.Errorf("write field %s: %w", key, err)
		}
	}

	for _, file := range files {
		if err := c.appendFile(writer, file.Field, file.Path); err != nil {
			writer.Close()
			return err
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(method), &body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call telegram %s: %w", method, err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("telegram %s failed: status=%d body=%s", method, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var apiResp tgResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return fmt.Errorf("decode telegram response: %w", err)
	}
	if !apiResp.OK {
		return fmt.Errorf("telegram %s error: %s", method, apiResp.Description)
	}

	return nil
}

func (c *Client) appendFile(writer *multipart.Writer, field, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file %s: %w", path, err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile(field, filepath.Base(path))
	if err != nil {
		return fmt.Errorf("create form file %s: %w", field, err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy file %s: %w", path, err)
	}
	return nil
}

func (c *Client) endpoint(method string) string {
	return fmt.Sprintf("%s/%s", c.apiURL, strings.TrimLeft(method, "/"))
}

type inlineKeyboardMarkup struct {
	InlineKeyboard [][]inlineKeyboardButton `json:"inline_keyboard"`
}

type inlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}
