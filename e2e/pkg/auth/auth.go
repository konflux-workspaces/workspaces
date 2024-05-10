package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
)

func BuildJwtForUser(ctx context.Context, user string) (string, error) {
	c := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveWorkspacesNamespace(ctx)

	s := corev1.Secret{}
	k := types.NamespacedName{Namespace: ns, Name: "workspaces-traefik-jwt-keys"}
	if err := c.Client.Get(ctx, k, &s); err != nil {
		return "", err
	}

	block, _ := pem.Decode(s.Data["private"])
	parseResult, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	key, ok := parseResult.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("Failed to retrieve RSA private key")
	}

	return jwt.NewWithClaims(jwt.SigningMethodRS512,
		jwt.MapClaims{
			"iss": "e2e-test",
			"sub": user,
		}).SignedString(key)
}
