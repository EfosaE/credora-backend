package authsvc

import (
	"context"
	"net/http"
	"time"

	"github.com/EfosaE/credora-backend/domain/auth"
	"github.com/EfosaE/credora-backend/internal/response"
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
	_, tokenString, err := j.tokenAuth.Encode(map[string]interface{}{
		"user_id":        payload.UserID,
		"name":           payload.Name,
		"account_number": payload.AccountNumber,
		"iat":            time.Now().Unix(),
		"exp":            time.Now().Add(j.expiryDur).Unix(),
	})

	return tokenString, err
}

// // Complete the ParseToken method
// func (j *JWTTokenService) ParseToken(ctx context.Context, tokenString string) (auth.TokenPayload, error) {
// 	token, err := j.tokenAuth.Decode(tokenString)
// 	if err != nil {
// 		return auth.TokenPayload{}, err
// 	}

// 	claims, ok := token.Claims.(map[string]interface{})
// 	if !ok || !token.Valid {
// 		return auth.TokenPayload{}, fmt.Errorf("invalid token")
// 	}

// 	// Extract claims
// 	userIDStr, ok := claims["user_id"].(string)
// 	if !ok {
// 		return auth.TokenPayload{}, fmt.Errorf("invalid user_id claim")
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		return auth.TokenPayload{}, fmt.Errorf("invalid user_id format")
// 	}

// 	name, _ := claims["name"].(string)
// 	accountNumber, _ := claims["account_number"].(string)

// 	return auth.TokenPayload{
// 		UserID:        userID,
// 		Name:          name,
// 		AccountNumber: accountNumber,
// 	}, nil
// }

// func (j *JWTTokenService) GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
// 	// Refresh tokens typically have longer expiry
// 	refreshExpiry := time.Hour * 24 * 7 // 7 days

// 	_, tokenString, err := j.tokenAuth.Encode(map[string]interface{}{
// 		"user_id": userID,
// 		"type":    "refresh",
// 		"iat":     time.Now().Unix(),
// 		"exp":     time.Now().Add(refreshExpiry).Unix(),
// 	})
// 	return tokenString, err
// }

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
