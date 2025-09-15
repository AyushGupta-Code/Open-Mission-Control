package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var verifier *oidc.IDTokenVerifier

func InitOIDC() error {
	issuer := os.Getenv("KEYCLOAK_ISSUER")
	if issuer == "" {
		issuer = "http://localhost:8081/realms/open-mission-control"
	}

	provider, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	config := &oidc.Config{
		ClientID: os.Getenv("KEYCLOAK_CLIENT_ID"),
	}

	verifier = provider.Verifier(config)
	return nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		idToken, err := verifier.Verify(r.Context(), tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Save claims in context (so handlers can use them)
		var claims map[string]interface{}
		if err := idToken.Claims(&claims); err == nil {
			ctx := context.WithValue(r.Context(), "user", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
		}
	})
}
