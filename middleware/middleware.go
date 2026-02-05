package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/nikhilpratapgit/TodoApp/database/dbHelper"
	"github.com/nikhilpratapgit/TodoApp/models"
)

type userContextKeyType struct{}

var userContextKey = userContextKeyType{}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		sessionUUID, err := uuid.Parse(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		userID, err := dbHelper.ValidateSession(sessionUUID.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		user := &models.UserCtx{
			UserID:    userID.String(),
			SessionID: sessionUUID.String(),
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserContext(r *http.Request) *models.UserCtx {
	user, _ := r.Context().Value(userContextKey).(*models.UserCtx)
	return user
}
