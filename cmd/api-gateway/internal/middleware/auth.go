package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

// func Auth(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.Header.Get("Authorization") == "" {
// 			http.Error(w, "unauthorized", http.StatusUnauthorized)
// 			return
// 		}

// 		// ✅ Respect client-provided role (curl / gateway / future JWT)
// 		if r.Header.Get("X-Role") == "" {
// 			r.Header.Set("X-Role", "USER") // sensible default
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer"))
		jwtSecret := os.Getenv("JWT_SECRET")
		jwtRequired := strings.EqualFold(os.Getenv("JWT_REQUIRED"), "true")

		if jwtSecret != "" {
			claims, err := verifyHS256(token, jwtSecret)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			if role, _ := claims["role"].(string); role != "" {
				r.Header.Set("X-Role", strings.ToUpper(role))
			}
		} else if jwtRequired {
			http.Error(w, "jwt required", http.StatusUnauthorized)
			return
		}

		if r.Header.Get("X-Role") == "" {
			r.Header.Set("X-Role", "USER")
		}

		next.ServeHTTP(w, r)
	})
}

func verifyHS256(token, secret string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errInvalidToken
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expectedSig := mac.Sum(nil)

	gotSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(gotSig, expectedSig) {
		return nil, errInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errInvalidToken
	}

	now := time.Now().Unix()
	if exp, ok := claims["exp"].(float64); ok && int64(exp) < now {
		return nil, errInvalidToken
	}
	if nbf, ok := claims["nbf"].(float64); ok && int64(nbf) > now {
		return nil, errInvalidToken
	}

	return claims, nil
}

var errInvalidToken = &tokenError{"invalid jwt"}

type tokenError struct{ msg string }

func (e *tokenError) Error() string { return e.msg }
