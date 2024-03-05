package auth

import (
	"context"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"

	rcontext "github.com/konflux-workspaces/workspaces/server/core/context"
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
	if err != nil {
		if _, err := w.Write([]byte(err.Error())); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := tkn.Claims.GetSubject()
	if err != nil {
		if _, err := w.Write([]byte(err.Error())); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, rcontext.UserKey, u)

	p.next.ServeHTTP(w, r.WithContext(ctx))
}
