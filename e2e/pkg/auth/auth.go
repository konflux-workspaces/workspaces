package auth

import "github.com/golang-jwt/jwt/v5"

func BuildJwtForUser(user string) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": "e2e-test",
			"sub": user,
		})
}
