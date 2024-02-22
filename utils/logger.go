package utils

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/TwiN/go-color"
	"github.com/google/uuid"
)

const (
	VIVIAN_LOGGER_SUCCESS      string = "[vivian:success]"
	VIVIAN_LOGGER_DEBUG        string = "[vivian:debug]"
	VIVIAN_LOGGER_WARNING      string = "[vivian:warning]"
	VIVIAN_LOGGER_ERROR        string = "[vivian:error]"
	VIVIAN_LOGGER_FATAL        string = "[vivian:fatal]"
	HTTP_PROTOCOL_DEPLOYMENT   string = "\033[35m"
	SOCKET_PROTOCOL_DEPLOYMENT string = "\033[36m"
)

type VivianLogger struct {
	Logger       *log.Logger
	DeploymentID string
	Protocol     uint16
}

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
		deploymentProtocol = color.Ize(SOCKET_PROTOCOL_DEPLOYMENT, s.DeploymentID[:8])
	default:
		deploymentProtocol = color.Ize(HTTP_PROTOCOL_DEPLOYMENT, s.DeploymentID[:8])
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

func generateDeploymentID() string {
	randomUUID := uuid.New()
	shortUUID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		randomUUID[:4], randomUUID[4:6], randomUUID[6:8],
		randomUUID[8:10], randomUUID[10:])

	return shortUUID
}

func (s *VivianLogger) Deploy(statusDB bool) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	projectName := filepath.Base(wd)

	deploymentID := generateDeploymentID()
	s.DeploymentID = deploymentID

	fmt.Printf("╭───────────────────────────────────────────────────╮\n")
	fmt.Printf("│ app        : %-45s │\n", color.Ize(color.Cyan, projectName))
	fmt.Printf("│ deployment : %-36s │\n", color.Ize(color.Purple, deploymentID))
	fmt.Printf("╰───────────────────────────────────────────────────╯\n")
}

func (s *VivianLogger) DeployLong(statusDB bool) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	projectName := filepath.Base(wd)

	deploymentID := generateDeploymentID()
	s.DeploymentID = deploymentID

	fmt.Printf("╭───────────────────────────────────────────────────╮\n")
	fmt.Printf("│ app        : %-45s │\n", color.Ize(color.Cyan, projectName))
	//fmt.Printf("│ database   : %-45s │\n", color.Ize(color.Green, fmt.Sprintf("status:%v", statusDB)))
	fmt.Printf("│ deployment : %-36s │\n", color.Ize(color.Purple, deploymentID))
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
