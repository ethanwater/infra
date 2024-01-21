package main

import (
	"context"

	"vivian.infra/internal/app"
)

func main() {
	err := app.Deploy(context.Background())
	if err != nil {
		return
	}

}
