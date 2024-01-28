package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"vivian.infra/database"
)

func fetchUserAccount(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		alias := vars["alias"]

		account, err := database.FetchAccount(VivianDatabase, alias)
		if err != nil {
			VivianServerLogger.LogWarning(fmt.Sprintf("unable to fetch account %v", err))
			return
		}

		VivianServerLogger.LogSuccess(fmt.Sprintf("fetched account: {%v %v}", account.ID, account.Email))
	})
}
