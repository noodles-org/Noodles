package middleware

import (
	"context"
	"net/http"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"github.com/mephalrith/noodles/backend/internal/config"
	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/model"
	"github.com/mephalrith/noodles/backend/internal/respond"
	"github.com/mephalrith/noodles/backend/internal/services"
)

type contextKey string

const userContextKey contextKey = "user"

func UserFromContext(ctx context.Context) *model.User {
	u, _ := ctx.Value(userContextKey).(*model.User)
	return u
}

func RequireAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.IsProduction {
				ctx := context.WithValue(r.Context(), userContextKey, &model.DevUser)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			cookie, err := r.Cookie(cfg.JWT.CookieName)
			if err != nil {
				services.AuthEvents.With(map[string]string{"status": "failure", "reason": "no_token"}).Inc()
				services.Logger.Warn("Auth: missing token", "ip", r.RemoteAddr, "path", r.URL.Path)
				respond.Error(w, errs.Unauthorized)
				return
			}

			token, err := jwtlib.Parse(cookie.Value, func(t *jwtlib.Token) (any, error) {
				return []byte(cfg.JWT.Secret), nil
			})
			if err != nil || !token.Valid {
				services.AuthEvents.With(map[string]string{"status": "failure", "reason": "invalid_token"}).Inc()
				services.Logger.Warn("Auth: invalid token", "ip", r.RemoteAddr, "error", err)
				respond.Error(w, errs.InvalidToken)
				return
			}

			claims, ok := token.Claims.(jwtlib.MapClaims)
			if !ok {
				respond.Error(w, errs.InvalidClaims)
				return
			}

			var groups []string
			if g, ok := claims["groups"]; ok {
				if gs, ok := g.([]any); ok {
					for _, v := range gs {
						if s, ok := v.(string); ok {
							groups = append(groups, s)
						}
					}
				}
			}

			user := &model.User{
				Sub:    claimString(claims, "sub"),
				Email:  claimString(claims, "email"),
				Name:   claimString(claims, "name"),
				Role:   model.Role(claimString(claims, "role")),
				Groups: groups,
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...model.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				respond.Error(w, errs.Forbidden)
				return
			}

			for _, role := range roles {
				if user.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			services.Logger.Warn("Auth: insufficient role",
				"user", user.Email, "role", user.Role, "required", roles)
			respond.Error(w, errs.Forbidden)
		})
	}
}

func claimString(claims jwtlib.MapClaims, key string) string {
	if v, ok := claims[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
