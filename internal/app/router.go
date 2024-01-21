package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"vivian.infra/internal/pkg/auth"
)

const (
	BUCKET_LIMITER_SIZE uint16 = 10
	BUCKET_LIMITER_LEAK_RATE uint16 = 1
)

var requestFlag atomic.Uint32
var requestChannel chan uint16 = make(chan uint16, BUCKET_LIMITER_SIZE)
var wg sync.WaitGroup

//TODO: fix this, use a ticker to take away leak_rate, <- chan
func init() {
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case requestChannel <- 1:
		default:
		}
	}()
}

func authentication2FA(ctx context.Context, server *Server) http.Handler {
	wg.Wait()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Channel Len:", len(requestChannel), "Channel Cap:", cap(requestChannel))
		fmt.Println("Channel Len:", requestFlag.Load())
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
			requestChannel <- 1
			generateAuthentication2FA(w, ctx, server)
		case "verify":
			key := strings.TrimSpace(q.Get("key"))
			verifyAuthentication2FA(w, ctx, server, key)
		case "expire":
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
