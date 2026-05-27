package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type updateResponse struct {
	OK     bool     `json:"ok"`
	Result []update `json:"result"`
}

type update struct {
	UpdateID int64    `json:"update_id"`
	Message  *message `json:"message"`
}

type message struct {
	Chat chat   `json:"chat"`
	Text string `json:"text"`
}

type chat struct {
	ID int64 `json:"id"`
}

type bot struct {
	token     string
	webAppURL string
	client    *http.Client
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	webAppURL := os.Getenv("TELEGRAM_WEB_APP_URL")
	if token == "" || webAppURL == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN and TELEGRAM_WEB_APP_URL are required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	b := &bot{
		token:     token,
		webAppURL: webAppURL,
		client:    &http.Client{Timeout: 35 * time.Second},
	}

	if err := b.setMenuButton(ctx); err != nil {
		log.Printf("set menu button: %v", err)
	}

	log.Println("telegram bot started")
	if err := b.poll(ctx); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
}

func (b *bot) poll(ctx context.Context) error {
	var offset int64
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updates, err := b.getUpdates(ctx, offset)
		if err != nil {
			log.Printf("get updates: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, upd := range updates {
			offset = upd.UpdateID + 1
			if upd.Message == nil || upd.Message.Text != "/start" {
				continue
			}
			if err := b.sendWelcome(ctx, upd.Message.Chat.ID); err != nil {
				log.Printf("send welcome: %v", err)
			}
		}
	}
}

func (b *bot) getUpdates(ctx context.Context, offset int64) ([]update, error) {
	reqBody := map[string]any{
		"offset":  offset,
		"timeout": 25,
	}
	var result updateResponse
	if err := b.call(ctx, "getUpdates", reqBody, &result); err != nil {
		return nil, err
	}
	if !result.OK {
		return nil, fmt.Errorf("telegram returned ok=false")
	}
	return result.Result, nil
}

func (b *bot) sendWelcome(ctx context.Context, chatID int64) error {
	body := map[string]any{
		"chat_id": chatID,
		"text":    "Открой семейное древо в Web App.",
		"reply_markup": map[string]any{
			"inline_keyboard": [][]map[string]any{
				{
					{
						"text": "Открыть древо",
						"web_app": map[string]string{
							"url": b.webAppURL,
						},
					},
				},
			},
		},
	}
	return b.call(ctx, "sendMessage", body, nil)
}

func (b *bot) setMenuButton(ctx context.Context) error {
	body := map[string]any{
		"menu_button": map[string]any{
			"type": "web_app",
			"text": "Family Tree",
			"web_app": map[string]string{
				"url": b.webAppURL,
			},
		},
	}
	return b.call(ctx, "setChatMenuButton", body, nil)
}

func (b *bot) call(ctx context.Context, method string, payload any, out any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.apiURL(method), bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("telegram status %d: %s", res.StatusCode, string(body))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(body, out)
}

func (b *bot) apiURL(method string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.token, method)
}
