package logging

import (
	"context"
	"log/slog"

	"github.com/cidverse/cidverseutils/zerologconfig"
)

// Trace calls [Logger.Trace] on the default logger.
func Trace(msg string, args ...any) {
	slog.Default().Log(context.Background(), zerologconfig.LevelTrace, msg, args...)
}
