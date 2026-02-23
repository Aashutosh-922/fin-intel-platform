package middleware

import "net/http"

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
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Respect incoming role if provided
		if r.Header.Get("X-Role") == "" {
			r.Header.Set("X-Role", "USER")
		}

		next.ServeHTTP(w, r)
	})
}
