package logging

import (
	"io"
	"log/slog"
	"os"
	"sonarbridge-go/configs"
)

var logger = slog.Default()

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Set(l *slog.Logger) {
	logger = l
}

func InitServer(cfg configs.Config) error {
	lv, err := GetLevel(cfg)
	if err != nil {
		return nil
	}

	logger = slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: lv,
			}))
	return nil
}

func InitCLI(cfg configs.Config) error {
	lvl, err := GetLevel(cfg)
	if err != nil {
		return err
	}

	logger = slog.New(
		slog.NewTextHandler(
			os.Stderr,
			&slog.HandlerOptions{
				Level: lvl,
			},
		),
	)
	return nil
}

func NewWithWriter(
	writer io.Writer,
	cfg configs.Config,
	json bool,
) (*slog.Logger, error) {

	lvl, err := GetLevel(cfg)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		Level: lvl,
	}

	if json {
		return slog.New(slog.NewJSONHandler(writer, opts)), nil
	}

	return slog.New(slog.NewTextHandler(writer, opts)), nil
}

func SetOutput(w io.Writer) {
	logger = slog.New(
		slog.NewTextHandler(w, nil))
}
