package filter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sunweiwe/horizon/pkg/apiserver/request"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
)

func WithRequestInfo(next http.Handler, resolver request.RequestInfoResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		authorization := req.Header.Get("Authorization")
		if len(authorization) == 0 {
			xAuthorization := req.Header.Get("X-Horizon-Authorization")
			if len(xAuthorization) != 0 {
				req.Header.Set("Authorization", xAuthorization)
				req.Header.Del("X-Horizon-Authorization")
			}
		}

		if len(req.URL.Query()["dryrun"]) != 0 {
			req.URL.RawQuery = strings.Replace(req.URL.RawQuery, "dryrun", "dryRun", 1)
		}

		if rawQuery := req.Header.Get("X-Horizon-RawQuery"); len(rawQuery) != 0 && len(req.URL.RawQuery) == 0 {
			req.URL.RawQuery = rawQuery
			req.Header.Del("X-Horizon-RawQuery")
		}

		ctx := req.Context()
		info, err := resolver.NewRequestInfo(req)
		if err != nil {
			responsewriters.InternalError(w, req, fmt.Errorf("Failed to create RequestInfo: %v", err))
			return
		}

		req = req.WithContext(request.WithRequestInfo(ctx, info))
		next.ServeHTTP(w, req)
	})
}
