package middleware

import (
	"log"
	"net/http"

	"github.com/Temisaputra/warOnk/pkg/auth"
	"github.com/Temisaputra/warOnk/pkg/helper"
	"github.com/prometheus/client_golang/prometheus"
)

type AuthMiddleware struct {
	jwtSvc  auth.JwtService
	counter prometheus.Counter
}

func NewAuthMiddleware(jwtSvc auth.JwtService, counter prometheus.Counter) *AuthMiddleware {
	return &AuthMiddleware{jwtSvc: jwtSvc, counter: counter}
}

func (a *AuthMiddleware) Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user, err := a.jwtSvc.ValidateCurrentUser(r)
		if err != nil {
			log.Println("Error validating user:", err)
			helper.WriteResponse(w, err, nil)
			return
		}

		a.counter.Inc()
		// simpan user ke context
		r = auth.SetUserContext(r, user)
		next.ServeHTTP(w, r)
	})
}
