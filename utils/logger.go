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

type VivianLogger struct {
	ProjectName  string
	Logger       *log.Logger
	LogFile      string //assigned log file
	DeploymentID string //shortened UUID
	Protocol     uint16 // http: 0 | websocket: 1
	mux          sync.Mutex
}

func (s *VivianLogger) Deploy(databaseConnectionStatus bool) {
	deploymentUUID := uuid.New()
	s.DeploymentID = fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		deploymentUUID[:4], deploymentUUID[4:6], deploymentUUID[6:8],
		deploymentUUID[8:10], deploymentUUID[10:])

	workingDirectory, err := os.Getwd()
	if err != nil {
		s.ProjectName = "null"
	} else {
		s.ProjectName = filepath.Base(workingDirectory)
	}

	_, err = os.Stat("logs")
	if os.IsNotExist(err) {
		if err := os.Mkdir("logs", os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	if len(s.LogFile) <= 0 {
		currentTime := time.Now()
		month := currentTime.Month()

		_, err = os.Stat(fmt.Sprintf("%s/%s", "logs", month))
		if os.IsNotExist(err) {
			if err := os.Mkdir(fmt.Sprintf("%s/%s", "logs/", month), os.ModePerm); err != nil {
				log.Fatal(err)
			}
		}

		logFileHeader := fmt.Sprintf("logs/%s/%d-%v.log",
			month,
			currentTime.Day(),
			s.DeploymentID,
		)

		f, err := os.Create(logFileHeader)
		if err != nil {
			return
		}
		defer f.Close() //blocks if the logs directory doesnt exist
		s.LogFile = logFileHeader
	}

	fmt.Printf("╭───────────────────────────────────────────────────╮\n│ app        : %-45s │\n│ deployment : %-36s │\n", color.Ize(color.Cyan, s.ProjectName), color.Ize(color.Purple, s.DeploymentID))
	if databaseConnectionStatus {
		fmt.Printf("│ database : %-36s │\n", color.Ize(color.Purple, s.DeploymentID))
	}
	fmt.Printf("╰───────────────────────────────────────────────────╯\n")
}

func (s *VivianLogger) logMessage(logLevel, msg string, isErrorMessage bool) {
	invocationTime := time.Now().UTC().Format("2006-01-02 15:04:05")
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		fmt.Println("failed to get file information")
		return
	}
	file = path.Base(file)

	var deploymentProtocol string
	switch s.Protocol {
	case 1:
		deploymentProtocol = color.Ize(SOCKET_PROTOCOL_DEPLOYMENT, s.DeploymentID[:8])
	default:
		deploymentProtocol = color.Ize(HTTP_PROTOCOL_DEPLOYMENT, s.DeploymentID[:8])
	}

	s.mux.Lock()
	defer s.mux.Unlock()
	logFileMessage := fmt.Sprintf(
		"%v %-35s %s %s",
		invocationTime,
		fmt.Sprintf("%s:%v:", file, line),
		s.DeploymentID[:8],
		msg,
	)
	s.logToFile(logFileMessage)

	if isErrorMessage {
		msg = color.Ize(color.Yellow, msg)
	}

	logServerMessage := fmt.Sprintf(
		"%v %-35s %s %-25s %s",
		invocationTime,
		color.Ize(color.Blue, fmt.Sprintf("%s:%v:", file, line)),
		deploymentProtocol,
		logLevel,
		msg,
	)
	s.Logger.Print(logServerMessage)

}
func (s *VivianLogger) logToFile(msg string) error {
	log, err := os.OpenFile(s.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer log.Close()
	log.WriteString(fmt.Sprintf("%v\n", msg))
	return nil
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

func (s *VivianLogger) SetProtocol(protocol uint16) {
	s.Protocol = protocol
}

func (s *VivianLogger) DefaultProtocol() {
	s.Protocol = 0
}
