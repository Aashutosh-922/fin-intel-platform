package middleware

import "net/http"

func RequireRole(allowed ...string) func(http.Handler) http.Handler {
	allowedSet := make(map[string]struct{})
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Header.Get("X-Role")
			if role == "" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			if _, ok := allowedSet[role]; !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
