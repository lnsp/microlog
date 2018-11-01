package tokens

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

const (
	emailExpiresAfter   = 24 * time.Hour
	sessionExpiresAfter = time.Hour
)

type EmailPurpose string

const (
	PurposeReset        EmailPurpose = "password_reset"
	PurposeConfirmation              = "email_confirmation"
)

type EmailTokenClaims struct {
	jwt.StandardClaims
	ID      uint
	Email   string
	Purpose EmailPurpose
}

func CreateEmailToken(secret []byte, email string, uid uint, purpose EmailPurpose) (string, error) {
	claims := EmailTokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(emailExpiresAfter).Unix(),
		},
		ID:      uid,
		Email:   email,
		Purpose: purpose,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to create session")
	}
	return signedToken, nil
}

func VerifyEmailToken(secret []byte, signedToken string, purpose EmailPurpose) (string, uint, bool) {
	var claims EmailTokenClaims
	_, err := jwt.ParseWithClaims(signedToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return "", 0, false
	}
	if claims.Purpose != purpose {
		return "", 0, false
	}
	return claims.Email, claims.ID, true
}

type SessionClaims struct {
	jwt.StandardClaims
	Username  string
	ID        uint
	Moderator bool
}

func VerifySessionToken(secret []byte, signedToken string) (string, uint, bool, bool) {
	var claims SessionClaims
	_, err := jwt.ParseWithClaims(signedToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return "", 0, false, false
	}
	return claims.Username, claims.ID, claims.Moderator, true
}

func CreateSessionToken(secret []byte, name string, id uint, mod bool) (string, error) {
	claims := SessionClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(sessionExpiresAfter).Unix(),
		},
		Username:  name,
		ID:        id,
		Moderator: mod,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to create session")
	}
	return signedToken, nil
}
