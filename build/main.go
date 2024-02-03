package main

import (
	"context"
	"os"
	"runtime/pprof"

	"vivian.infra/internal/app"
)

func main() {
	f, _ := os.Create("profile.pprof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	
	err := app.Deploy(context.Background()); if err != nil {
		return 
	}
}
