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
	"vivian.infra/database"
	"vivian.infra/utils"
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
	router := mux.NewRouter()
	router.Handle("/{alias}/2FA", authentication2FA(ctx)).Methods("GET")
	router.Handle("/{alias}/fetch", fetchUserAccount(ctx)).Methods("GET")
	router.Handle("/ws", HandleWebSocketTimestamp(ctx))

	deploymentID := generateDeploymentID()
	vivianServer := &Server{
		DeploymentID:       deploymentID,
		Logger:             &utils.VivianLogger{Logger: log.New(os.Stdout, "", log.Lmsgprefix), DeploymentID: deploymentID},
		Handler:            router,
		Addr:               VIVIAN_HOST_ADDR,
		VivianReadTimeout:  VIVIAN_READWRITE_TIMEOUT,
		VivianWriteTimeout: VIVIAN_READWRITE_TIMEOUT,
	}
	VivianServerLogger = vivianServer.Logger

	vivianServer.mux.Lock()
	configSQL := database.ConfigSQL{
		Driver: "mysql",
		Source: "root:@tcp(127.0.0.1:3306)/user_schema",
	}
	err := configSQL.InitDatabase(ctx, VivianServerLogger)
	if err != nil {
		VivianServerLogger.LogError("unable to connect to SQL database", err)
	}
	VivianDatabase = configSQL.Database
	vivianServer.mux.Unlock()

	httpServer := &http.Server{
		Addr:         vivianServer.Addr,
		Handler:      vivianServer.Handler,
		ReadTimeout:  vivianServer.VivianReadTimeout,
		WriteTimeout: vivianServer.VivianWriteTimeout,
	}
	vivianServer.Logger.LogDeployment(VivianDatabase.Ping() == nil)

	go func() {
		<-ctx.Done()
		httpServer.Shutdown(context.Background())
	}()

	return http.ListenAndServe(vivianServer.Addr, vivianServer.Handler)
}

func generateDeploymentID() string {
	randomUUID := uuid.New()
	shortUUID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		randomUUID[:4], randomUUID[4:6], randomUUID[6:8],
		randomUUID[8:10], randomUUID[10:])

	return shortUUID
}
