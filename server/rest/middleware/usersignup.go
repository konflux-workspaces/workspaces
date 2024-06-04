package middleware

import (
	"context"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/cache"

	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
)

type UserSignupMiddleware struct {
	cache             cache.Cache
	requireUserSignup bool

	next http.Handler
}

func NewUserSignupMiddleware(next http.Handler, cache cache.Cache, requireUserSignup bool) *UserSignupMiddleware {
	return &UserSignupMiddleware{
		cache:             cache,
		requireUserSignup: requireUserSignup,

		next: next,
	}
}

func (m *UserSignupMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// retrieve User's JWT Sub
	u, ok := r.Context().Value(ccontext.UserSubKey).(string)
	if !ok {
		m.next.ServeHTTP(w, r)
		return
	}

	// retrieve UserSignup for given sub
	us, err := m.lookupUserSignup(r.Context(), u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if us == nil {
		w.WriteHeader(http.StatusForbidden)
		if _, err := w.Write([]byte("user needs to sign in")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// user is waiting for approval
	if us.Status.CompliantUsername == "" {
		w.WriteHeader(http.StatusForbidden)
		if _, err := w.Write([]byte("user is waiting for approval")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// TODO(@filariow): check if user is deactivated or banned

	// inject the userSignup.ComplaintUsername
	ctx := context.WithValue(r.Context(), ccontext.UserSignupComplaintNameKey, us.Status.CompliantUsername)
	m.next.ServeHTTP(w, r.WithContext(ctx))
}

func (m *UserSignupMiddleware) lookupUserSignup(ctx context.Context, sub string) (*toolchainv1alpha1.UserSignup, error) {
	uu := toolchainv1alpha1.UserSignupList{}
	if err := m.cache.List(ctx, &uu); err != nil {
		return nil, err
	}

	for _, u := range uu.Items {
		if u.Spec.IdentityClaims.Sub == sub {
			return &u, nil
		}
	}

	return nil, nil
}
