package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"vivian.infra/internal/pkg/auth"
)

func authentication2FA(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*RequestChannelCounter++

		q := r.URL.Query()
		action := strings.TrimSpace(q.Get("action"))
		// curl "localhost:8080/bella/2FA?action="
		switch action {
		case "generate":
			*RequestChannel <- 1
			generateAuthentication2FA(w, ctx)
		case "verify":
			*RequestChannel <- 1
			key := strings.TrimSpace(q.Get("key"))
			verifyAuthentication2FA(w, ctx, key)
		case "expire":
			*RequestChannel <- 1
			expireAuthentication2FA(w, ctx)
		default:
			http.NotFound(w, r)
		}
	})
}

func generateAuthentication2FA(w http.ResponseWriter, ctx context.Context) {
	keyChan := make(chan string)
	errorChan := make(chan error)

	go func() {
		key2FA, err := auth.GenerateAuthKey2FA(ctx, VivianServerLogger)
		if err != nil {
			errorChan <- err
			return
		}
		keyChan <- key2FA
	}()

	select {
	case hash2FA := <-keyChan:
		_, err := json.Marshal(hash2FA)
		if err != nil {
			VivianServerLogger.LogError("failure marshalling results", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//if _, err := fmt.Fprintln(w, "2FA generation successful"); err != nil {
		//	VivianServerLogger.LogError("Failure writing results", err)
		//	return
		//}
	case err := <-errorChan:
		VivianServerLogger.LogError("unable to generate authentication 2FA: %v", err)
		return
	}
}

func verifyAuthentication2FA(w http.ResponseWriter, ctx context.Context, key2FA string) {
	resultChan := make(chan bool)
	errorChan := make(chan error)

	go func() {
		result, err := auth.VerifyAuthKey2FA(ctx, key2FA, VivianServerLogger)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
		KillRequestTickerChannel <- 1
	}()

	select {
	case result := <-resultChan:
		bytes, err := json.Marshal(result)
		if err != nil {
			VivianServerLogger.LogError("failure marshalling results", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
			VivianServerLogger.LogError("failure writing results", err)
			return
		}
	case err := <-errorChan:
		VivianServerLogger.LogError("unable to verify key", errors.New("invalid Key"))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func expireAuthentication2FA(w http.ResponseWriter, ctx context.Context) {
	err := auth.Expire2FA(ctx, VivianServerLogger)
	if err != nil {
		VivianServerLogger.LogError("failed to expire 2FA ->", err)
		return
	}
	VivianServerLogger.LogSuccess("successfully expired 2FA token")
}
