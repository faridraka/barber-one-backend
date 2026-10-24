package middleware

import (
	"context"
	"net/http"
)

func AuthMiddleware(next http.Handler, sessionStore *SessionStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := sessionStore.UserIDFromRequest(r)
		if !ok {
			WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Silakan login untuk mengakses endpoint ini."})
			return
		}
		
		// Add user ID to context
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
