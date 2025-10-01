package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type postBody struct {
	Username  string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
	Content   string `json:"content"`
}

type DiscordClient struct {
	logger     *slog.Logger
	Client     *http.Client
	WebHookURL string
	Username   string
	AvatarUrl  string
}

func NewDiscordClient(logger *slog.Logger, webHookURL string, timeout time.Duration, username string, avatarUrl string) *DiscordClient {
	return &DiscordClient{
		logger:     logger,
		WebHookURL: webHookURL,
		Client: &http.Client{
			Timeout: timeout,
		},
		Username:  username,
		AvatarUrl: avatarUrl,
	}
}

func (d *DiscordClient) Post(ctx context.Context, content string) error {
	post := postBody{
		Username:  d.Username,
		AvatarUrl: d.AvatarUrl,
		Content:   content,
	}

	body, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to marshal post: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.WebHookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil

}
