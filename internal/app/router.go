package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"vivian.infra/internal/pkg/auth"
)

func EchoResponseHandler(ctx context.Context, server *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//TODO validate if user exists and is valid
		//vars := mux.Vars(r)
		//user := vars["user"]
		//if user does not exist{
		//	logWarning*
		//	return
		//}

		vars := mux.Vars(r)
		echoResponse := vars["echo"]

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			server.Logger.LogSuccess(echoResponse)
		}()
		wg.Wait()

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(echoResponse))
		if err != nil {
			server.Logger.LogError("Error writing response: ", err)
		}
	})
}

func Authentication2FA(ctx context.Context, server *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//TODO validate if user exists and is valid
		//vars := mux.Vars(r)
		//user := vars["user"]
		//if user does not exist{
		//	logWarning*
		//	return
		//}

		q := r.URL.Query()
		action := strings.TrimSpace(q.Get("action"))
		switch action {
		case "generate":
			GenerateAuthentication2FA(w, ctx, server)
		case "verify":
			key := strings.TrimSpace(q.Get("key"))
			VerifyAuthentication2FA(w, ctx, server, key)
		case "expire":
			ExpireAuthentication2FA(w, ctx, server)
		default:
			http.NotFound(w, r)
		}
	})
}

func GenerateAuthentication2FA(w http.ResponseWriter, ctx context.Context, server *Server) {
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
			server.Logger.LogError("failure marshalling results", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
			server.Logger.LogError("failure writing results", err)
			return
		}
	case err := <-errorChan:
		server.Logger.LogError("unable to generate authentication 2FA: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func VerifyAuthentication2FA(w http.ResponseWriter, ctx context.Context, server *Server, key2FA string) {
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
			server.Logger.LogError("failure marshalling results", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
			server.Logger.LogError("failure writing results", err)
			return
		}
	case err := <-errorChan:
		//server.Logger.LogWarning("invalid key")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ExpireAuthentication2FA(w http.ResponseWriter, ctx context.Context, server *Server) {
	err := auth.Expire2FA(ctx, server.Logger)
	if err != nil {
		server.Logger.LogError("failed to expire 2FA ->", err)
		return
	}
	server.Logger.LogDebug("successfully expired 2FA token")
}
