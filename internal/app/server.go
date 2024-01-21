package app

import (
	"context"
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

func Deploy(ctx context.Context) error {
	logger := log.New(os.Stdout, "", log.Lmsgprefix)
	s := buildServer(ctx, logger)
	s.mux.Lock()
	defer s.mux.Unlock()

	server := &http.Server{
		Addr:         s.Addr,
		Handler:      s.Handler,
		ReadTimeout:  s.VivianReadTimeout,
		WriteTimeout: s.VivianWriteTimeout,
	}

	s.Logger.LogDeployment()

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
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
	server := &Server{
		DeploymentID:       deploymentID,
		Logger:             &utils.VivianLogger{Logger: logger, DeploymentID: deploymentID},
		Handler:            router,
		Addr:               VIVIAN_HOST_ADDR,
		VivianReadTimeout:  VIVIAN_READWRITE_TIMEOUT,
		VivianWriteTimeout: VIVIAN_READWRITE_TIMEOUT,
	}

	//2FA handlers are called onyl after the user is verified via Login
	router.Handle("/{user}/2FA", authentication2FA(ctx, server)).Methods("GET")
	return server
}
