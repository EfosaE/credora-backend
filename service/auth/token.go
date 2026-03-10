package authsvc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/EfosaE/credora-backend/domain/auth"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/go-chi/jwtauth/v5"
)

type JWTTokenService struct {
	secret    string
	expiryDur time.Duration
	tokenAuth *jwtauth.JWTAuth
}

func NewJWTTokenService(secret string, expiry time.Duration) *JWTTokenService {
	tokenAuth := jwtauth.New("HS256", []byte(secret), nil)
	return &JWTTokenService{secret: secret, expiryDur: expiry, tokenAuth: tokenAuth}
}

func (j *JWTTokenService) GenerateToken(ctx context.Context, payload auth.TokenPayload) (string, error) {
	utils.PrintJSON(payload)
	_, tokenString, err := j.tokenAuth.Encode(map[string]any{
		"userId":        payload.UserID,
		"name":          payload.Name,
		"iat":           time.Now().Unix(),
		"exp":           time.Now().Add(j.expiryDur).Unix(),
	})

	return tokenString, err
}

// HTTP middleware methods
func (j *JWTTokenService) Verifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(j.tokenAuth)
}

func (j *JWTTokenService) Authenticator() func(http.Handler) http.Handler {
	// return jwtauth.Authenticator(j.tokenAuth)
	return CustomAuthenticator(j.tokenAuth)
}

// Authenticator is a default authentication middleware to enforce access from the
// Verifier middleware request context values. The Authenticator sends a 401 Unauthorized
// response for any unverified tokens and passes the good ones through. It's just fine
// until you decide to write something similar and customize your client response.
func CustomAuthenticator(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := jwtauth.FromContext(r.Context())

			if err != nil {
				response.SendError(w, r, response.Unauthorized(err.Error()))
				// http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if token == nil {
				response.SendError(w, r, response.Unauthorized(http.StatusText(http.StatusUnauthorized)))
				return
			}

			// Token is authenticated, pass it through
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func GenerateResetToken() (rawToken string, hashedToken string, expiresAt time.Time, err error) {
	// 1. Generate random bytes
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}

	// 2. Encode as hex string (same as Node)
	rawToken = hex.EncodeToString(b)

	// 3. Hash token using bcrypt (cost 12)
	hash, err := bcrypt.GenerateFromPassword([]byte(rawToken), 12)
	if err != nil {
		return
	}
	hashedToken = string(hash)

	// 4. Expiration time (1 hour)
	expiresAt = time.Now().Add(1 * time.Hour)

	return
}
