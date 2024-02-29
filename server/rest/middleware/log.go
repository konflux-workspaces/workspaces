package middleware

import (
	"log"
	"net/http"
)

var _ http.Handler = &LogRequestMiddleware{}

type LogRequestMiddleware struct {
	next http.Handler
}

func NewLogRequestMiddleware(next http.Handler) *LogRequestMiddleware {
	return &LogRequestMiddleware{next: next}
}

func (p *LogRequestMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.String())
	p.next.ServeHTTP(w, r)
}
