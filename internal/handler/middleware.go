package handler

import (
	"net/http"

	"github.com/atnomoverflow/auth-server/pkg/helpers/response"
	"github.com/atnomoverflow/auth-server/pkg/token"
)

func ValiateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := token.ExtractToken(r)
		if err != nil {
			msg := err.Error()
			response.UnauthorizedError(w, &msg)
			return
		}
		// save token in context much eazier to extract
		next.ServeHTTP(w, r)
	})
}
