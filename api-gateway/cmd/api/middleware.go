package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
)

type Claims struct {
	UserID int32  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type contextKey string

const claimsKey contextKey = "claims"

func (app *Config) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Unauthorized(w, "Missing auth header")
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			response.Unauthorized(w, "Invalid auth header")
			return
		}

		if headerParts[0] != "Bearer" {
			response.Unauthorized(w, "Invalid auth header")
			return
		}

		token := headerParts[1]
		resp, err := app.ValidateTokenViaGRPC(r.Context(), token)
		if err != nil || !resp.Valid {
			response.Unauthorized(w, "Invalid token:")
			return
		}
		userClaims := &Claims{
			UserID: resp.UserId,
			Email:  resp.Email,
			Role:   resp.Role,
		}
		ctx := context.WithValue(r.Context(), claimsKey, userClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (app *Config) GetClaims(ctx context.Context) (*Claims, error) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	if !ok {
		return nil, errors.New("no claims found in context")
	}
	return claims, nil
}
