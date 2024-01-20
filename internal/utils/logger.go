package utils

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/TwiN/go-color"
)

const (
	VivianAppName       string = "vivian.infra"
	VivianLoggerSuccess string = "[vivian:success]"
	VivianLoggerDebug 	string = "[vivian:debug]"
	VivianLoggerWarning string = "[vivian:warn]"
	VivianLoggerError   string = "[vivian:error]"
	VivianLoggerFatal   string = "[vivian:fatal]"
)

type VivianLogger struct {
	Logger       *log.Logger
	DeploymentID string
}

func (s *VivianLogger) logMessage(logLevel, msg string) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		fmt.Println("Failed to get file information")
		return
	}

	filename := path.Base(file)
	logMessage := fmt.Sprintf(
		"%v %-35s %s %-25s %s",
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		color.Ize(color.Blue, fmt.Sprintf("%s:%v:", filename, line)),
		color.Ize(color.Purple, s.DeploymentID[:8]),
		logLevel,
		msg,
	)

	s.Logger.Print(logMessage)
}

func (s *VivianLogger) LogDeployment() {
	fmt.Printf("╭───────────────────────────────────────────────────╮\n")
	fmt.Printf("│ app        : %-45s │\n", color.Ize(color.Cyan, VivianAppName))
	fmt.Printf("│ deployment : %-36s │\n", color.Ize(color.Purple, s.DeploymentID))
	fmt.Printf("╰───────────────────────────────────────────────────╯\n")
}

func (s *VivianLogger) LogSuccess(msg string) {
	s.logMessage(color.Ize(color.Green, VivianLoggerSuccess), msg)
}

func (s *VivianLogger) LogDebug(msg string) {
	s.logMessage(color.Ize(color.Cyan, VivianLoggerDebug), msg)
}

func (s *VivianLogger) LogWarning(msg string) {
	s.logMessage(color.Ize(color.Yellow, VivianLoggerWarning), msg)
}

func (s *VivianLogger) LogError(msg string, err error) {
	s.logMessage(color.Ize(color.Red, VivianLoggerError), color.Ize(color.Yellow, fmt.Sprintf("%s error: %s", msg, err)))
}

func (s *VivianLogger) LogFatal(msg string) {
	s.logMessage(color.Ize(color.RedBackground, VivianLoggerFatal), msg)
	os.Exit(1)
}
