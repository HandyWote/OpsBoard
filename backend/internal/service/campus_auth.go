package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"backend/internal/config"

	"go.uber.org/zap"
)

type campusAuthenticator struct {
	cfg    config.CampusAuthConfig
	client *http.Client
	log    *zap.Logger
}

func newCampusAuthenticator(cfg config.CampusAuthConfig, log *zap.Logger) campusVerifier {
	if !cfg.Enabled || strings.TrimSpace(cfg.LoginURL) == "" {
		return nil
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	if log == nil {
		log = zap.NewNop()
	}

	return &campusAuthenticator{
		cfg: cfg,
		client: &http.Client{
			Timeout: timeout,
		},
		log: log,
	}
}

func (c *campusAuthenticator) Verify(ctx context.Context, username, password string) error {
	form := url.Values{}
	form.Set("opr", c.cfg.Operation)
	form.Set("userName", username)
	form.Set("pwd", password)
	form.Set("rememberPwd", c.cfg.Remember)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.LoginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return err
	}
	bodyText := strings.TrimSpace(string(body))

	if resp.StatusCode >= 500 {
		return fmt.Errorf("campus gateway unavailable: %s", resp.Status)
	}

	parsed, msg, parseErr := parseCampusResponse(bodyText)
	if parseErr != nil {
		c.log.Warn("无法解析校园网认证返回", zap.Error(parseErr), zap.String("body", truncateForLog(bodyText, 200)))
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("campus gateway error: %s", resp.Status)
		}
		return fmt.Errorf("campus gateway unexpected response")
	}

	if !parsed {
		if msg == "" {
			msg = "校园网认证失败"
		}
		return fmt.Errorf("%w: %s", ErrInvalidCredentials, msg)
	}

	return nil
}

func parseCampusResponse(body string) (bool, string, error) {
	if body == "" {
		return false, "", fmt.Errorf("empty response")
	}

	// 优先尝试直接解析 JSON
	var payload struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal([]byte(body), &payload); err == nil {
		if payload.Msg == "" {
			payload.Msg = payload.Message
		}
		return payload.Success, payload.Msg, nil
	}

	normalized := strings.ReplaceAll(body, "'", "\"")
	if err := json.Unmarshal([]byte(normalized), &payload); err == nil {
		if payload.Msg == "" {
			payload.Msg = payload.Message
		}
		return payload.Success, payload.Msg, nil
	}

	if strings.Contains(strings.ToLower(body), "logon success") {
		return true, "logon success", nil
	}
	if strings.Contains(strings.ToLower(body), "password error") {
		return false, "密码错误", nil
	}
	return false, "", fmt.Errorf("unrecognized campus response")
}

func truncateForLog(s string, limit int) string {
	if limit <= 0 || len(s) <= limit {
		return s
	}
	return s[:limit] + "..."
}
