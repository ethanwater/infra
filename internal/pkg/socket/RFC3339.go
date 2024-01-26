package socket

import (
	"context"
	"time"
)

type T interface {
	LoggerSocket(context.Context, string) error
}

func Time(ctx context.Context) ([]byte, error) {
	return []byte(time.Now().Format(time.RFC3339)), nil
}
