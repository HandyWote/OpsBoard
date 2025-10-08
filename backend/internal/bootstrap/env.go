package bootstrap

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

// loadDotEnv 读取简单的 KEY=VALUE 形式的环境文件。
func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		if key == "" {
			continue
		}
		value := strings.TrimSpace(line[idx+1:])
		value = strings.Trim(value, "\"'")
		_ = os.Setenv(key, value)
	}
	return scanner.Err()
}
