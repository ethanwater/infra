package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	_ "embed"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"vivian.infra/internal/database"
	"vivian.infra/internal/utils"
)

const (
	VIVIAN_APP_NAME          string        = "vivian.infra"
	VIVIAN_HOST_ADDR         string        = ":8080"
	VIVIAN_READWRITE_TIMEOUT time.Duration = time.Second * 10
)

type ServerInitialization interface {
	Deploy(context.Context) error
}

type Server struct {
	DeploymentID       string
	Listener           net.Listener
	Handler            http.Handler
	Logger             *utils.VivianLogger
	Addr               string
	VivianReadTimeout  time.Duration
	VivianWriteTimeout time.Duration
	mux                sync.Mutex
}

var (
	VivianServerLogger *utils.VivianLogger
	VivianDatabase     *sql.DB
)

func Deploy(ctx context.Context) error {
	s := buildServer(ctx, log.New(os.Stdout, "", log.Lmsgprefix))
	VivianServerLogger = s.Logger

	configSQL := database.ConfigSQL{
		Driver: "mysql",
		Source: "root:@tcp(127.0.0.1:3306)/",
	}
	err := configSQL.InitDatabase(ctx, VivianServerLogger)
	if err != nil {
		VivianServerLogger.LogFatal("unable to connect to SQL database")
	}
	VivianDatabase = configSQL.Database

	s.mux.Lock()
	defer s.mux.Unlock()
	httpServer := &http.Server{
		Addr:         s.Addr,
		Handler:      s.Handler,
		ReadTimeout:  s.VivianReadTimeout,
		WriteTimeout: s.VivianWriteTimeout,
	}
	s.Logger.LogDeployment(VivianDatabase.Ping() == nil)

	go func() {
		<-ctx.Done()
		httpServer.Shutdown(context.Background())
	}()

	return http.ListenAndServe(s.Addr, s.Handler)
}

func buildServer(ctx context.Context, logger *log.Logger) *Server {
	generateDeploymentID := func() string {
		randomUUID := uuid.New()
		shortUUID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			randomUUID[:4], randomUUID[4:6], randomUUID[6:8],
			randomUUID[8:10], randomUUID[10:])

		return shortUUID
	}
	deploymentID := generateDeploymentID()

	router := mux.NewRouter()
	vivianServer := &Server{
		DeploymentID:       deploymentID,
		Logger:             &utils.VivianLogger{Logger: logger, DeploymentID: deploymentID},
		Handler:            router,
		Addr:               VIVIAN_HOST_ADDR,
		VivianReadTimeout:  VIVIAN_READWRITE_TIMEOUT,
		VivianWriteTimeout: VIVIAN_READWRITE_TIMEOUT,
	}

	//2FA handlers are called onyl after the user is verified via Login
	router.Handle("/{user}/2FA", authentication2FA(ctx, vivianServer)).Methods("GET")
	router.Handle("/{user}/fetch", fetchUserAccount(ctx)).Methods("GET")
	return vivianServer
}
