package auth

// HTTP integration

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "userID"

// Retrieves the authenticated user's ID from the request context
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

// Contains HTTP authentication middleware
type Middleware struct {
	jwt *JWTManager
}

// Creates authentication middleware using the supplied JWT manager
func NewMiddleware(jwtManager *JWTManager) *Middleware {
	return &Middleware{jwt: jwtManager}
}

// Verifies the JWT in the Authorization header before allowing
// the request to reach the handler
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Error unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(header, " ", 2)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "Error unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(parts[1])

		if token == "" {
			http.Error(w, "Error unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := m.jwt.Parse(token)
		if err != nil {
			http.Error(w, "Error unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
