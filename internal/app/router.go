package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"vivian.infra/internal/pkg/auth"
)

const (
	BUCKET_LIMITER_SIZE      uint32        = 10
	BUCKET_LIMITER_LEAK_AMT  uint32        = 1
	BUCKET_LIMITER_LEAK_RATE time.Duration = 500 * time.Millisecond
)

var requestChannel *chan uint32
var killRequestTickerChannel = make(chan uint16)
var requestChannelCounter *uint32

type Limiter struct {
	requestTicker         time.Ticker
	requestChannel        chan uint32
	requestChannelCounter uint32
	requestBlockerState   atomic.Uint32
}

func init() {
	//TODO: dont intitiate counter if app is not in 2FA state
	limiter := Limiter{
		requestTicker:         *time.NewTicker(BUCKET_LIMITER_LEAK_RATE),
		requestChannel:        make(chan uint32, BUCKET_LIMITER_SIZE),
		requestChannelCounter: 0,
		requestBlockerState:   atomic.Uint32{},
	}
	requestChannel = &limiter.requestChannel
	requestChannelCounter = &limiter.requestChannelCounter
	limiter.RateLimiter()
}

// TODO: should have its own file
func (l *Limiter) RateLimiter() {
	go func() {
		for {
			select {
			case <-l.requestTicker.C:
				if l.requestChannelCounter >= 10 {
					VivianServerLogger.LogWarning(fmt.Sprintf("Blocking channel {status code:%v}", http.StatusTooManyRequests))
				}
				fmt.Println("Channel Len:", len(l.requestChannel), "Channel Cap:", cap(l.requestChannel), "Pool:", l.requestChannelCounter)
				if l.requestChannelCounter > 0 {
					l.requestChannelCounter -= BUCKET_LIMITER_LEAK_AMT
				} else {
					l.requestChannelCounter = 0
					l.requestChannel = make(chan uint32, BUCKET_LIMITER_SIZE)
				}
			case <-killRequestTickerChannel:
				VivianServerLogger.LogDebug("killing request ticker channel")
				return
			}
		}
	}()
}

func authentication2FA(ctx context.Context, server *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *requestChannelCounter >= 10 {
			return
		}
		*requestChannelCounter++
		//TODO validate if user exists and is valid
		//vars := mux.Vars(r)
		//detect user session***
		//user := vars["user"]
		//if user does not exist{
		//	logWarning*
		//	return
		//}

		q := r.URL.Query()
		action := strings.TrimSpace(q.Get("action"))
		switch action {
		case "generate":
			*requestChannel <- 1
			generateAuthentication2FA(w, ctx, server)
		case "verify":
			*requestChannel <- 1
			key := strings.TrimSpace(q.Get("key"))
			verifyAuthentication2FA(w, ctx, server, key)
		case "expire":
			*requestChannel <- 1
			expireAuthentication2FA(w, ctx, server)
		default:
			http.NotFound(w, r)
		}
	})
}

func generateAuthentication2FA(w http.ResponseWriter, ctx context.Context, server *Server) {
	keyChan := make(chan string)
	errorChan := make(chan error)

	go func() {
		key2FA, err := auth.GenerateAuthKey2FA(ctx, server.Logger)
		if err != nil {
			errorChan <- err
			return
		}
		keyChan <- key2FA
	}()

	select {
	case hash2FA := <-keyChan:
		bytes, err := json.Marshal(hash2FA)
		if err != nil {
			server.Logger.LogError("Failure marshalling results", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
			server.Logger.LogError("Failure writing results", err)
			return
		}
	case err := <-errorChan:
		server.Logger.LogError("Unable to generate authentication 2FA: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func verifyAuthentication2FA(w http.ResponseWriter, ctx context.Context, server *Server, key2FA string) {
	resultChan := make(chan bool)
	errorChan := make(chan error)

	go func() {
		result, err := auth.VerifyAuthKey2FA(ctx, key2FA, server.Logger)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
		killRequestTickerChannel <- 1
	}()

	select {
	case result := <-resultChan:
		bytes, err := json.Marshal(result)
		if err != nil {
			server.Logger.LogError("Failure marshalling results", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
			server.Logger.LogError("Failure writing results", err)
			return
		}
	case err := <-errorChan:
		server.Logger.LogError("Unable to verify key", errors.New("invalid Key"))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func expireAuthentication2FA(w http.ResponseWriter, ctx context.Context, server *Server) {
	err := auth.Expire2FA(ctx, server.Logger)
	if err != nil {
		server.Logger.LogError("Failed to expire 2FA ->", err)
		return
	}
	server.Logger.LogSuccess("Successfully expired 2FA token")
}
