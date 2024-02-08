package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var _ http.Handler = &JwtBearerMiddleware{}

type JwtBearerMiddleware struct {
	next http.Handler
}

func NewJwtBearerMiddleware(next http.Handler) *JwtBearerMiddleware {
	return &JwtBearerMiddleware{next: next}
}

func (p *JwtBearerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jp := jwt.NewParser()
	a := r.Header.Get("Authorization")
	t := strings.TrimPrefix(a, "Bearer ")
	if t == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// TODO: Verify Signature
	tkn, _, err := jp.ParseUnverified(t, jwt.MapClaims{})
	u, err := tkn.Claims.GetSubject()
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, "user", u)

	p.next.ServeHTTP(w, r.WithContext(ctx))
}
