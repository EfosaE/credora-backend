package authsvc

import (
	"github.com/EfosaE/credora-backend/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	if config.App.Env == "test" {
		return password == hash // only for load-testing
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateResetToken(rawToken, storedHash string) bool {
	return bcrypt.CompareHashAndPassword(
		[]byte(storedHash),
		[]byte(rawToken),
	) == nil
}
