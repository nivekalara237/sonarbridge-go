package pkg

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sonarbridge-go/internal/infra/utils"
	"sonarbridge-go/pkg/numbers"
	stringify "sonarbridge-go/pkg/string"
	"strings"
)

type Environ interface {
	Get(key string) (any, error)
	GetOrDefault(key string, orElse any) any
	GetString(key string) (string, error)
	GetInt(key string) (int64, error)
	GetBool(key string) (bool, error)
}

type DotEnv struct {
	dotenvPath string
}

var envs map[string]any

func New(path string) *DotEnv {
	return &DotEnv{dotenvPath: path}
}

func (d *DotEnv) Load() (map[string]any, error) {
	path := d.dotenvPath
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return map[string]any{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("dotenv: read %s: %w", path, err)
	}
	defer f.Close()

	env := map[string]any{}
	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		compile := regexp.MustCompile(`^(.*?)\s#(.*)`)
		matches := compile.FindStringSubmatch(text)
		if len(matches) > 1 {
			beforeHash := matches[1]
			text = strings.TrimSpace(beforeHash)
		}

		key, value, ok := strings.Cut(text, "=")
		if !ok {
			return nil, fmt.Errorf("dotenv: %s:%d: expected KEY=VALUE, got %q", path, line, text)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		env[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("dotenv: read %s: %w", path, err)
	}
	envs = env
	return envs, err
}

func (d *DotEnv) Get(key string) (any, error) {
	v, ok := envs[key]
	if !ok {
		return nil, fmt.Errorf("dotenv: %s not found", key)
	}
	return v, nil
}

func (d *DotEnv) GetOrDefault[T any](key string, orElse T) T {
	v, err := d.Get(key)
	if err != nil {
		return orElse
	}
	return v.(T)
}

func (d *DotEnv) GetString(key string) (string, error) {
	v, ok := envs[key]
	if !ok {
		return "", fmt.Errorf("dotenv: %s not found", key)
	}
	return stringify.ToString(v), nil
}

func (d *DotEnv) GetInt(key string) (int64, error) {
	v, ok := envs[key]
	if !ok {
		return 0, fmt.Errorf("dotenv: %s not found", key)
	}
	return numbers.ToInt64(v)
}

func (d *DotEnv) GetBool(key string) (bool, error) {
	v, ok := envs[key]
	if !ok {
		return false, fmt.Errorf("dotenv: %s not found", key)
	}
	switch b := v.(type) {
	case bool:
		return b, nil
	case string:
		s := strings.ToLower(strings.TrimSpace(b))
		if slices.Contains([]string{"true", "false"}, s) {
			return utils.Ternary(s == "true", true, false), nil
		}
		return false, fmt.Errorf("cannot convert %T to bool", key)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", key)
	}
}
