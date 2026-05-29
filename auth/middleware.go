package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	Sub         string   `json:"sub"`
	Aud         []string `json:"aud"`
	Iss         string   `json:"iss"`
	Email       string   `json:"email,omitempty"`
	Name        string   `json:"name,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

type contextKey string

const UserContextKey contextKey = "user"

var jwks *keyfunc.JWKS

func InitJWKS() error {
	jwksURL := "https://pleasant-bengal-78.clerk.accounts.dev/.well-known/jwks.json"

	var err error
	jwks, err = keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshErrorHandler: func(err error) {
			fmt.Println("JWKS refresh error:", err)
		},
	})

	return err
}

func JWTMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("ENV") == "local" {
			userId := "user_3BsFWuuSolrouYsrV3h9thejXIX"
			ctx := context.WithValue(r.Context(), UserContextKey, userId)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, jwks.Keyfunc)
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}

		// Clerk: just extract the user ID (sub claim)
		userId, ok := claims["sub"].(string)
		if !ok || userId == "" {
			http.Error(w, "User ID (sub) claim is missing", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserContextKey, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(UserContextKey).(string)
	if !ok || user == "" {
		return "", false
	}
	return user, true
}
