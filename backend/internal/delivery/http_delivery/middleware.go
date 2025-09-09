package http_delivery

import (
	"context"
	"encoding/json"
	"net/http"
	"real-time-chat/internal/services"
	"real-time-chat/internal/usecase"
	"strings"
)

type contextKey string

const userContextKey contextKey = "user"

func AuthMiddleware(jwtSecret string, userService usecase.UserUseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				ErrorResponse(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
				ErrorResponse(w, http.StatusUnauthorized, "Invalid Authorization header format")
				return
			}
			accessToken := tokenParts[1]

			// A real app should use a dedicated token service here,
			// but for now, we access it via the userService's internal tokenService dependency.
			// This is a slight architectural compromise for simplicity given the tool's constraints.
			claims, err := userService.(*services.UserService).TokenService.ValidateToken(accessToken)
			if err != nil {
				ErrorResponse(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			userID := claims.UserID
			user, err := userService.GetUserByID(r.Context(), userID)
			if err != nil {
				ErrorResponse(w, http.StatusUnauthorized, "User not found")
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	JSONResponse(w, statusCode, map[string]string{"error": message})
}
