package session

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type UserInfo struct {
	Identity uint32
	Name     string
	Role     string
}

type Claims struct {
	jwt.StandardClaims
	UserInfo
}

func (svc *Server) GenerateToken(info *UserInfo) (string, error) {
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(svc.expiration).Unix(),
		},
		UserInfo: *info,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(svc.secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate token")
	}
	return signed, nil
}

func (svc *Server) ProofToken(signed string) (*UserInfo, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(signed, &claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return svc.secret, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse token")
	}
	return &claims.UserInfo, nil
}
