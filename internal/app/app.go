package app

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
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
	DeploymentID       *string
	Listener           net.Listener
	Handler            http.Handler
	Logger             *utils.VivianLogger
	Addr               string
	VivianReadTimeout  time.Duration
	VivianWriteTimeout time.Duration
}

var (
	VivianServerLogger *utils.VivianLogger
)

func Deploy(ctx context.Context) error {
	router := mux.NewRouter()

	vivianServer := &Server{
		Logger:             &utils.VivianLogger{Logger: log.New(os.Stdout, "", log.Lmsgprefix)},
		Handler:            router,
		Addr:               VIVIAN_HOST_ADDR,
		VivianReadTimeout:  VIVIAN_READWRITE_TIMEOUT,
		VivianWriteTimeout: VIVIAN_READWRITE_TIMEOUT,
	}
	VivianServerLogger = vivianServer.Logger
	vivianServer.Logger.Deploy(false)

	//router.Handle("/{alias}/fetch", fetchUserAccount(ctx)).Methods("GET")
	router.Handle("/{alias}/2FA", authentication2FA(ctx)).Methods("GET")
	router.Handle("/sockettime", HandleWebSocketTimestamp(ctx))
	router.Handle("/{alias}/bucket/fetch", fetchBucketContents()).Methods("GET")

	httpServer := &http.Server{
		Addr:         vivianServer.Addr,
		Handler:      vivianServer.Handler,
		ReadTimeout:  vivianServer.VivianReadTimeout,
		WriteTimeout: vivianServer.VivianWriteTimeout,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			vivianServer.Logger.LogError("server error", err)
		}
	}()

	<-ctx.Done()

	if err := httpServer.Close(); err != nil {
		vivianServer.Logger.LogError("shutdown error", err)
	}

	return nil
}
