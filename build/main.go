package main

import (
	"context"
	"os"

	"vivian.infra/internal/app"
)

func main() {
	err := app.Deploy(context.Background()); if err != nil {
		os.Exit(1)
	}
}
