package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"github.com/mephalrith/noodles/backend/internal/config"
	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/middleware"
	"github.com/mephalrith/noodles/backend/internal/model"
	"github.com/mephalrith/noodles/backend/internal/respond"
	"github.com/mephalrith/noodles/backend/internal/services"
)

// authFailure logs a warning or error, increments the failure metric, and redirects to the login page.
func authFailure(w http.ResponseWriter, r *http.Request, reason, redirect, msg string, args ...any) {
	services.Logger.Error(msg, args...)
	services.AuthEvents.With(map[string]string{"status": "failure", "reason": reason}).Inc()
	http.Redirect(w, r, "/login?error="+redirect, http.StatusFound)
}

func resolveRole(groups []string, cfg *config.Config) model.Role {
	for _, g := range groups {
		for _, ag := range cfg.Auth.AdminGroups {
			if g == ag {
				return model.RoleAdmin
			}
		}
	}
	return model.RoleViewer
}

func isAllowedUser(groups []string, cfg *config.Config) bool {
	for _, g := range groups {
		for _, ag := range cfg.Auth.AllowedGroups {
			if g == ag {
				return true
			}
		}
	}
	return false
}

// signSessionJWT creates a signed JWT string from a User.
func signSessionJWT(user *model.User, secret string) (string, error) {
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, jwtlib.MapClaims{
		"sub":    user.Sub,
		"email":  user.Email,
		"name":   user.Name,
		"role":   string(user.Role),
		"groups": user.Groups,
	})
	return token.SignedString([]byte(secret))
}

// setSessionCookie writes the dashboard JWT cookie to the response.
func setSessionCookie(w http.ResponseWriter, cfg *config.Config, signed string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.JWT.CookieName,
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.IsProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   8 * 60 * 60,
	})
}

func HandleLogin(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !cfg.IsProduction {
			signed, err := signSessionJWT(&model.DevUser, cfg.JWT.Secret)
			if err != nil {
				http.Error(w, "Failed to sign token", http.StatusInternalServerError)
				return
			}
			setSessionCookie(w, cfg, signed)
			services.Logger.Info("Auth: dev login bypass", "ip", r.RemoteAddr)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		b := make([]byte, 32)
		rand.Read(b)
		state := hex.EncodeToString(b)

		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			Path:     "/api/auth",
			HttpOnly: true,
			Secure:   cfg.IsProduction,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   600,
		})

		services.Logger.Info("Auth: login initiated", "ip", r.RemoteAddr)

		params := url.Values{
			"client_id":     {cfg.OAuth.ClientID},
			"redirect_uri":  {cfg.OAuth.CallbackURL},
			"response_type": {"code"},
			"scope":         {cfg.OAuth.Scopes},
			"state":         {state},
		}

		http.Redirect(w, r, fmt.Sprintf("%s?%s", cfg.OAuth.AuthorizeURL, params.Encode()), http.StatusFound)
	}
}

// exchangeToken exchanges an authorization code for an access token.
func exchangeToken(cfg *config.Config, code string) (string, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {cfg.OAuth.ClientID},
		"client_secret": {cfg.OAuth.ClientSecret},
		"code":          {code},
		"redirect_uri":  {cfg.OAuth.CallbackURL},
	}

	resp, err := http.Post(cfg.OAuth.TokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return "", errs.TokenRequestFailed.Wrap(err)
	}
	defer resp.Body.Close()

	var data struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", errs.TokenDecodeFailed.Wrap(err)
	}
	return data.AccessToken, nil
}

// fetchUserinfo calls the OAuth userinfo endpoint and returns the raw claims.
func fetchUserinfo(userinfoURL, accessToken string) (map[string]any, error) {
	req, err := http.NewRequest("GET", userinfoURL, nil)
	if err != nil {
		return nil, errs.UserinfoRequestFailed.Wrap(err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errs.UserinfoFetchFailed.Wrap(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errs.UserinfoReadFailed.Wrap(err)
	}

	var claims map[string]any
	if err := json.Unmarshal(body, &claims); err != nil {
		return nil, errs.UserinfoDecodeFailed.Wrap(err)
	}
	return claims, nil
}

// parseGroups extracts a string slice from the "groups" claim.
func parseGroups(claims map[string]any) []string {
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
	return groups
}

// claimStr returns a string value from a claims map, or empty string if missing/wrong type.
func claimStr(claims map[string]any, key string) string {
	if v, ok := claims[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// buildUserFromClaims constructs a User from raw userinfo claims and resolved role.
func buildUserFromClaims(claims map[string]any, groups []string, role model.Role) model.User {
	sub := claimStr(claims, "sub")
	if sub == "" {
		sub = claimStr(claims, "id")
	}
	email := claimStr(claims, "email")
	name := claimStr(claims, "name")
	if name == "" {
		name = claimStr(claims, "preferred_username")
	}
	if name == "" {
		name = email
	}
	return model.User{
		Sub:    sub,
		Email:  email,
		Name:   name,
		Groups: groups,
		Role:   role,
	}
}

func HandleCallback(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		oauthError := r.URL.Query().Get("error")

		storedState := ""
		if c, err := r.Cookie("oauth_state"); err == nil {
			storedState = c.Value
		}

		http.SetCookie(w, &http.Cookie{
			Name:   "oauth_state",
			Value:  "",
			Path:   "/api/auth",
			MaxAge: -1,
		})

		if oauthError != "" {
			authFailure(w, r, "oauth_error", "oauth_denied",
				"Auth: provider error", "error", oauthError, "ip", r.RemoteAddr)
			return
		}

		if state == "" || storedState == "" || state != storedState {
			authFailure(w, r, "invalid_state", "invalid_state",
				"Auth: state mismatch", "ip", r.RemoteAddr)
			return
		}

		if code == "" {
			authFailure(w, r, "no_code", "no_code",
				"Auth: no code returned", "ip", r.RemoteAddr)
			return
		}

		accessToken, err := exchangeToken(cfg, code)
		if err != nil {
			authFailure(w, r, "token_exchange", "auth_failed",
				"Auth: token exchange failed", "error", err, "ip", r.RemoteAddr)
			return
		}

		userinfo, err := fetchUserinfo(cfg.OAuth.UserinfoURL, accessToken)
		if err != nil {
			authFailure(w, r, "token_exchange", "auth_failed",
				"Auth: userinfo failed", "error", err, "ip", r.RemoteAddr)
			return
		}

		groups := parseGroups(userinfo)

		if !isAllowedUser(groups, cfg) {
			authFailure(w, r, "unauthorized_group", "not_authorized",
				"Auth: user not in allowed groups", "email", claimStr(userinfo, "email"), "groups", groups, "ip", r.RemoteAddr)
			return
		}

		user := buildUserFromClaims(userinfo, groups, resolveRole(groups, cfg))

		signed, err := signSessionJWT(&user, cfg.JWT.Secret)
		if err != nil {
			authFailure(w, r, "token_exchange", "auth_failed",
				"Auth: failed to sign JWT", "error", err)
			return
		}

		setSessionCookie(w, cfg, signed)
		services.TrackUniqueUser(user.Sub)
		services.Logger.Info("Auth: login success", "user", user.Email, "role", user.Role, "groups", groups, "ip", r.RemoteAddr)
		services.AuthEvents.With(map[string]string{"status": "success", "reason": "login"}).Inc()

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func HandleMe(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	respond.OK(w, user)
}

func HandleLogout(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.UserFromContext(r.Context())
		services.Logger.Info("Auth: logout", "user", user.Email, "ip", r.RemoteAddr)
		services.AuthEvents.With(map[string]string{"status": "success", "reason": "logout"}).Inc()

		http.SetCookie(w, &http.Cookie{
			Name:     cfg.JWT.CookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Expires:  time.Unix(0, 0),
		})

		respond.OK(w, map[string]bool{"ok": true})
	}
}
