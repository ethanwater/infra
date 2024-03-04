package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"vivian.infra/internal/pkg/s3"
)

func fetchBucketContents() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contents, err := s3.FetchBucketObjects()
		if err != nil {
			VivianServerLogger.LogWarning(fmt.Sprintf("unable to fetch contents%v", err))
			return
		}

		_, err = json.Marshal(contents)
		if err != nil {
			VivianServerLogger.LogError("failure marshalling results", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
		//	VivianServerLogger.LogError("failure writing results", err)
		//	return
		//}
		VivianServerLogger.LogSuccess(fmt.Sprintf("fetched contents: %v", contents))
	})
}
