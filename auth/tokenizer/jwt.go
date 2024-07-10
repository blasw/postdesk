package tokenizer

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

var (
	ErrTokenExpired = errors.New("token is expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

type Tokenizer interface {
	NewAccessToken(UserClaims) (string, error)
	NewRefreshToken() (string, error)
	ParseAccessToken(string) (*UserClaims, error)
	ParseRefreshToken(string) error
}

type JwtTokenizer struct {
	secret               string
	accessTokenLifetime  time.Duration
	refreshTokenLifetime time.Duration
}

type UserClaims struct {
	UserID uint64 `json:"user_id"`
	jwt.StandardClaims
}

func NewJwtTokenizer(secret string, accessTokenLifetime, refreshTokenLifetime time.Duration) *JwtTokenizer {
	return &JwtTokenizer{
		secret:               secret,
		accessTokenLifetime:  accessTokenLifetime,
		refreshTokenLifetime: refreshTokenLifetime,
	}
}

func (t *JwtTokenizer) NewAccessToken(userClaims UserClaims) (string, error) {
	userClaims.ExpiresAt = time.Now().Add(t.accessTokenLifetime).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	return token.SignedString([]byte(t.secret))
}

func (t *JwtTokenizer) NewRefreshToken() (string, error) {
	claims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(t.refreshTokenLifetime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.secret))
}

func (t *JwtTokenizer) validateToken(tokenString string) error {
	logrus.Debug("validating token")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what you expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logrus.Error("unexpected signing method: ", token.Header["alg"])
			return nil, ErrTokenInvalid
		}
		return []byte(t.secret), nil
	})

	if err != nil {
		logrus.WithError(err).Error("error occurred while trying to parse token")
		return ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		logrus.Debug("token is invalid")
		return ErrTokenInvalid
	}

	logrus.Debug("checking if token is expired")
	// Check if the token is expired
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return ErrTokenExpired
		}
	}

	return nil
}

func (t *JwtTokenizer) ParseAccessToken(accessToken string) (*UserClaims, error) {
	logrus.Debug("parsing access token: ", accessToken)
	err := t.validateToken(accessToken)
	if err != nil && errors.Is(err, ErrTokenInvalid) {
		logrus.Debug("access token is invalid")
		return nil, err
	}

	logrus.Debug("access token is valid or expired")

	claims := &UserClaims{}

	_, err = jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.secret), nil
	})
	if err != nil {
		logrus.WithError(err).Error("error occurred while trying to parse token")
		return nil, ErrTokenInvalid
	}

	logrus.Debug("Parsed successfully, claims: ", claims)

	return claims, nil
}

func (t *JwtTokenizer) ParseRefreshToken(refreshToken string) error {
	return t.validateToken(refreshToken)
}
