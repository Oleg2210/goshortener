package cookies

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const cookieUserID = "UserID"

type contextKey string

const contextUserID contextKey = cookieUserID

func sign(value string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func verify(value, signature string, secret []byte) bool {
	expected := sign(value, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func AuthMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieUserID)
			if err != nil {
				// cookie нет → создаём новую
				userID := uuid.NewString()
				sig := sign(userID, secret)

				http.SetCookie(w, &http.Cookie{
					Name:     cookieUserID,
					Value:    userID + "|" + sig,
					Path:     "/",
					HttpOnly: true,
				})

				ctx := context.WithValue(r.Context(), contextUserID, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// cookie есть → проверяем
			parts := strings.Split(cookie.Value, "|")
			if len(parts) != 2 || parts[0] == "" {
				// есть cookie, но она неправильная
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userID := parts[0]
			signature := parts[1]

			if !verify(userID, signature, secret) {
				// подпись неверная → выдаём новую
				newUserID := uuid.NewString()
				newSig := sign(newUserID, secret)

				http.SetCookie(w, &http.Cookie{
					Name:     cookieUserID,
					Value:    newUserID + "|" + newSig,
					Path:     "/",
					HttpOnly: true,
				})

				ctx := context.WithValue(r.Context(), contextUserID, newUserID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// всё ок → кладём userID в context
			ctx := context.WithValue(r.Context(), contextUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(contextUserID).(string)
	return id, ok
}
