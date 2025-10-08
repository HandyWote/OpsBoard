package service

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"backend/internal/config"

	"go.uber.org/zap"
)

type campusAuthenticator struct {
	loginURL string
	client   *http.Client
	log      *zap.Logger
}

func newCampusAuthenticator(cfg config.CampusAuthConfig, log *zap.Logger) campusVerifier {
	loginURL := strings.TrimSpace(cfg.LoginURL)
	if !cfg.Enabled || loginURL == "" {
		return nil
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	if log == nil {
		log = zap.NewNop()
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Warn("初始化 SSO CookieJar 失败，将在无会话情况下继续", zap.Error(err))
	}

	return &campusAuthenticator{
		loginURL: loginURL,
		client: &http.Client{
			Timeout: timeout,
			Jar:     jar,
		},
		log: log,
	}
}

func (c *campusAuthenticator) Verify(ctx context.Context, username, password string) error {
	loginPage, err := c.fetchLoginPage(ctx)
	if err != nil {
		return err
	}

	hidden := extractHiddenInputs(loginPage.Body)
	formAction := findFormAction(loginPage.Body)

	submitURL := loginPage.URL
	if formAction != "" {
		if parsed, parseErr := url.Parse(formAction); parseErr == nil {
			submitURL = loginPage.URL.ResolveReference(parsed)
		} else {
			c.log.Debug("解析 SSO form action 失败，回退到页面 URL", zap.Error(parseErr), zap.String("action", formAction))
		}
	}

	payload := url.Values{}
	for key, value := range hidden {
		payload.Set(key, value)
	}
	payload.Set("username", username)
	payload.Set("password", password)

	body, status, err := c.submitCredentials(ctx, submitURL.String(), payload.Encode())
	if err != nil {
		return err
	}

	if status >= http.StatusInternalServerError {
		return fmt.Errorf("campus sso unavailable: %d", status)
	}

	if detectSTUSuccess(body) {
		return nil
	}

	c.log.Debug("SSO 登录失败", zap.Int("status", status), zap.String("body_preview", truncateForLog(body, 256)))
	return fmt.Errorf("%w: 单点登录失败", ErrInvalidCredentials)
}

type ssoPage struct {
	URL  *url.URL
	Body string
}

func (c *campusAuthenticator) fetchLoginPage(ctx context.Context) (ssoPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.loginURL, nil)
	if err != nil {
		return ssoPage{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return ssoPage{}, err
	}
	defer resp.Body.Close()

	body, err := readBody(resp.Body, 512*1024)
	if err != nil {
		return ssoPage{}, err
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return ssoPage{}, fmt.Errorf("campus sso unavailable: %s", resp.Status)
	}

	pageURL := resp.Request.URL
	if pageURL == nil {
		if parsed, parseErr := url.Parse(c.loginURL); parseErr == nil {
			pageURL = parsed
		} else {
			return ssoPage{}, parseErr
		}
	}

	return ssoPage{
		URL:  pageURL,
		Body: body,
	}, nil
}

func (c *campusAuthenticator) submitCredentials(ctx context.Context, submitURL, formEncoded string) (string, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, submitURL, strings.NewReader(formEncoded))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, readErr := readBody(resp.Body, 512*1024)
	if readErr != nil {
		return "", resp.StatusCode, readErr
	}

	return body, resp.StatusCode, nil
}

func readBody(reader io.Reader, limit int64) (string, error) {
	if limit <= 0 {
		limit = 256 * 1024
	}
	data, err := io.ReadAll(io.LimitReader(reader, limit))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

var hiddenInputPattern = regexp.MustCompile(`(?is)<input\b[^>]*type\s*=\s*(?:'hidden'|"hidden"|hidden)[^>]*>`)

func extractHiddenInputs(markup string) map[string]string {
	result := make(map[string]string)
	matches := hiddenInputPattern.FindAllString(markup, -1)
	for _, tag := range matches {
		name, ok := extractAttr(tag, "name")
		if !ok || strings.TrimSpace(name) == "" {
			continue
		}
		value, _ := extractAttr(tag, "value")
		result[name] = value
	}
	return result
}

var formPattern = regexp.MustCompile(`(?is)<form\b[^>]*id\s*=\s*(?:'fm1'|"fm1")[^>]*>`)

func findFormAction(markup string) string {
	formIndex := formPattern.FindStringIndex(markup)
	if formIndex == nil {
		return ""
	}

	remaining := markup[formIndex[0]:]
	action, ok := extractAttr(remaining, "action")
	if ok {
		return action
	}
	return ""
}

func extractAttr(fragment, attr string) (string, bool) {
	doubleQuoted := regexp.MustCompile(`(?i)` + attr + `\s*=\s*"([^"]*)"`)
	if match := doubleQuoted.FindStringSubmatch(fragment); len(match) >= 2 {
		return html.UnescapeString(strings.TrimSpace(match[1])), true
	}

	singleQuoted := regexp.MustCompile(`(?i)` + attr + `\s*=\s*'([^']*)'`)
	if match := singleQuoted.FindStringSubmatch(fragment); len(match) >= 2 {
		return html.UnescapeString(strings.TrimSpace(match[1])), true
	}

	return "", false
}

func detectSTUSuccess(body string) bool {
	lower := strings.ToLower(body)
	if !strings.Contains(lower, "<frameset") {
		return false
	}

	keywords := []string{"banner.aspx", "index_menu.aspx", "page/extheadpage.aspx"}
	for _, kw := range keywords {
		if !strings.Contains(lower, kw) {
			return false
		}
	}
	return true
}

func truncateForLog(s string, limit int) string {
	if limit <= 0 || len(s) <= limit {
		return s
	}
	return s[:limit] + "..."
}
