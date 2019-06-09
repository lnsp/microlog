package session

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// UserInfo stores identifying information about a session user.
type UserInfo struct {
	Identity uint32
	Role     string
}

// Claims store user information in a JWT compatible way.
type Claims struct {
	jwt.StandardClaims
	UserInfo
}

// GenerateToken generates a new token based on the given user info.
func (s *Server) GenerateToken(info *UserInfo) (string, error) {
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.expiration).Unix(),
		},
		UserInfo: *info,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate token")
	}
	return signed, nil
}

// ProofToken checks if the given string is a valid token.
// If the token is valid, the retrieved user information will be returned.
func (s *Server) ProofToken(signed string) (*UserInfo, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(signed, &claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse token")
	}
	return &claims.UserInfo, nil
}
