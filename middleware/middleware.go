package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt"
	"github.com/nikhilpratapgit/TodoApp/database/dbHelper"
	"github.com/nikhilpratapgit/TodoApp/models"
	"github.com/nikhilpratapgit/TodoApp/utils"
)

type userContextKeyType struct{}

var userContextKey = userContextKeyType{}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		token, parseErr := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})
		if parseErr != nil || !token.Valid {
			utils.RespondError(w, http.StatusUnauthorized, parseErr, "invalid token")
			return
		}
		claimValues, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			utils.RespondError(w, http.StatusUnauthorized, nil, "invalid token claims")
			return
		}
		sessionID := claimValues["sessionId"].(string)
		archivedAt, err := dbHelper.GetArchivedAt(sessionID)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "internal server error")
			return
		}
		if archivedAt != nil {
			utils.RespondError(w, http.StatusUnauthorized, nil, "invalid token")
			return
		}

		//userID, err := dbHelper.ValidateSession(sessionID)
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusUnauthorized)
		//	return
		//}

		user := &models.UserCtx{
			UserID:    claimValues["userId"].(string),
			SessionID: sessionID,
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserContext(r *http.Request) *models.UserCtx {
	user, _ := r.Context().Value(userContextKey).(*models.UserCtx)
	return user
}
