package utils

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
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

var LogWriter string

type VivianLogger struct {
	ProjectName  string
	Logger       *log.Logger
	DeploymentID string
	Protocol     uint16
	Mux          sync.Mutex
}

func (s *VivianLogger) SetProtocol(protocol uint16) {
	s.Protocol = protocol
}

func (s *VivianLogger) DefaultProtocol() {
	s.Protocol = 0
}

func (s *VivianLogger) Deploy(statusDB bool) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	s.ProjectName = filepath.Base(wd)

	deploymentID := generateDeploymentID()
	s.DeploymentID = deploymentID

	currentTime := time.Now()
	logFileHeader := fmt.Sprintf("logs/%d-%d-%d-%v.log",
		currentTime.Year(),
		currentTime.Month(),
		currentTime.Day(),
		deploymentID,
	)

	f, err := os.Create(logFileHeader)
	if err != nil {
		return
	}
	defer f.Close()
	LogWriter = logFileHeader

	fmt.Printf("╭───────────────────────────────────────────────────╮\n")
	fmt.Printf("│ app        : %-45s │\n", color.Ize(color.Cyan, s.ProjectName))
	fmt.Printf("│ deployment : %-36s │\n", color.Ize(color.Purple, deploymentID))
	fmt.Printf("╰───────────────────────────────────────────────────╯\n")
}

func (s *VivianLogger) logMessage(logLevel, msg string, isError bool) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		fmt.Println("failed to get file information")
		return
	}

	displayMsg := msg
	if isError {
		displayMsg = color.Ize(color.Yellow, msg)
	}

	filename := path.Base(file)
	var deploymentProtocol string
	switch s.Protocol {
	case 1:
		deploymentProtocol = color.Ize(SOCKET_PROTOCOL_DEPLOYMENT, s.DeploymentID[:8])
	default:
		deploymentProtocol = color.Ize(HTTP_PROTOCOL_DEPLOYMENT, s.DeploymentID[:8])
	}

	currentTime := time.Now().UTC().Format("2006-01-02 15:04:05")
	displayLogMessage := fmt.Sprintf(
		"%v %-35s %s %-25s %s",
		currentTime,
		color.Ize(color.Blue, fmt.Sprintf("%s:%v:", filename, line)),
		deploymentProtocol,
		logLevel,
		displayMsg,
	)

	s.Mux.Lock()
	s.logToFile(msg, currentTime, filename, line)
	s.Mux.Unlock()

	s.Logger.Print(displayLogMessage)
}
func (s *VivianLogger) logToFile(msg string, time string, filename string, line int) {

	logMessage := fmt.Sprintf(
		"%v %-35s %s %s",
		time,
		fmt.Sprintf("%s:%v:", filename, line),
		s.DeploymentID[:8],
		msg,
	)

	log, err := os.OpenFile(LogWriter, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer log.Close()

	log.WriteString(fmt.Sprintf("%v\n", logMessage))
}

func (s *VivianLogger) LogSuccess(msg string) {
	s.logMessage(color.Ize(color.Green, VIVIAN_LOGGER_SUCCESS), msg, false)
}

func (s *VivianLogger) LogDebug(msg string) {
	s.logMessage(color.Ize(color.Gray, VIVIAN_LOGGER_DEBUG), msg, false)
}

func (s *VivianLogger) LogWarning(msg string) {
	s.logMessage(color.Ize(color.Yellow, VIVIAN_LOGGER_WARNING), msg, false)
}

func (s *VivianLogger) LogError(msg string, err error) {
	msg = fmt.Sprintf("%s error: %s", msg, err)
	s.logMessage(color.Ize(color.Red, VIVIAN_LOGGER_ERROR), msg, true)
}

func (s *VivianLogger) LogFatal(msg string) {
	s.logMessage(color.Ize(color.RedBackground, VIVIAN_LOGGER_FATAL), msg, false)
	os.Exit(1)
}

func generateDeploymentID() string {
	randomUUID := uuid.New()
	shortUUID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		randomUUID[:4], randomUUID[4:6], randomUUID[6:8],
		randomUUID[8:10], randomUUID[10:])

	return shortUUID
}
