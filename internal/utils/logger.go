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
	VIVIAN_APP_NAME       string = "vivian.infra"
	VIVIAN_LOGGER_SUCCESS string = "[vivian:success]"
	VIVIAN_LOGGER_DEBUG   string = "[vivian:debug]"
	VIVIAN_LOGGER_WARNING string = "[vivian:warning]"
	VIVIAN_LOGGER_ERROR   string = "[vivian:error]"
	VIVIAN_LOGGER_FATAL   string = "[vivian:fatal]"
)

type VivianLogger struct {
	Logger       *log.Logger
	DeploymentID string
	Protocol     uint16
}

var (
	httpProtocolDeployment   = color.Purple
	socketProtocolDeployment = color.Cyan
)

func (s *VivianLogger) logMessage(logLevel, msg string) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		fmt.Println("failed to get file information")
		return
	}

	filename := path.Base(file)
	var deploymentProtocol string
	switch s.Protocol {
	case 1:
		deploymentProtocol = color.Ize(socketProtocolDeployment, s.DeploymentID[:8])
	default:
		deploymentProtocol = color.Ize(httpProtocolDeployment, s.DeploymentID[:8])
	}

	logMessage := fmt.Sprintf(
		"%v %-35s %s %-25s %s",
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		color.Ize(color.Blue, fmt.Sprintf("%s:%v:", filename, line)),
		deploymentProtocol,
		logLevel,
		msg,
	)

	s.Logger.Print(logMessage)
}

func (s *VivianLogger) SetProtocol(protocol uint16) {
	s.Protocol = protocol
}

func (s *VivianLogger) DefaultProtocol() {
	s.Protocol = 0
}

func (s *VivianLogger) LogDeployment(statusDB bool) {
	fmt.Printf("╭───────────────────────────────────────────────────╮\n")
	fmt.Printf("│ app        : %-45s │\n", color.Ize(color.Cyan, VIVIAN_APP_NAME))
	fmt.Printf("│ database   : %-45s │\n", color.Ize(color.Green, fmt.Sprintf("status:%v", statusDB)))
	fmt.Printf("│ deployment : %-36s │\n", color.Ize(color.Purple, s.DeploymentID))
	fmt.Printf("╰───────────────────────────────────────────────────╯\n")
}

func (s *VivianLogger) LogSuccess(msg string) {
	s.logMessage(color.Ize(color.Green, VIVIAN_LOGGER_SUCCESS), msg)
}

func (s *VivianLogger) LogDebug(msg string) {
	s.logMessage(color.Ize(color.Gray, VIVIAN_LOGGER_DEBUG), msg)
}

func (s *VivianLogger) LogWarning(msg string) {
	s.logMessage(color.Ize(color.Yellow, VIVIAN_LOGGER_WARNING), msg)
}

func (s *VivianLogger) LogError(msg string, err error) {
	s.logMessage(color.Ize(color.Red, VIVIAN_LOGGER_ERROR), color.Ize(color.Yellow, fmt.Sprintf("%s error: %s", msg, err)))
}

func (s *VivianLogger) LogFatal(msg string) {
	s.logMessage(color.Ize(color.RedBackground, VIVIAN_LOGGER_FATAL), msg)
	os.Exit(1)
}
