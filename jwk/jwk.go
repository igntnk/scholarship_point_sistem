package jwk

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"time"
)

const (
	ClaimsContextKey = "claims"
)

type SPSAccessClaims struct {
	User   User           `json:"user"`
	Data   map[string]any `json:"data"`
	IsRoot bool
	jwt.RegisteredClaims
}

type SPSRefreshClaims struct {
	jwt.RegisteredClaims
	Data     map[string]any `json:"data"`
	UserUUID string         `json:"userUUID"`
}

type User struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	SecondName string `json:"second_name"`
	Patronymic string `json:"patronymic"`
	LastLogin  int64  `json:"last_login"`
}

type JWKExtractor interface {
	ParseAccessToken(tokenString string) (*SPSAccessClaims, error)
	ParseRefreshToken(refreshString string) (*SPSRefreshClaims, error)
}

type JWKSigner interface {
	JWKExtractor
	SignToken(claims jwt.Claims) (string, error)
	PublicKey() ([]byte, error)
}

type rsaJRS struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

func (s *rsaJRS) ParseAccessToken(tokenString string) (*SPSAccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &SPSAccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, errors.Join(err, unexpected.InternalErr)
	}

	if claims, ok := token.Claims.(*SPSAccessClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.Join(errors.New("invalid token"), unexpected.InternalErr)
}

func (s *rsaJRS) ParseRefreshToken(refreshString string) (*SPSRefreshClaims, error) {
	token, err := jwt.ParseWithClaims(refreshString, &SPSRefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	}, jwt.WithLeeway(time.Second*5))
	if err != nil {
		return nil, errors.Join(err, unexpected.InternalErr)
	}

	if claims, ok := token.Claims.(*SPSRefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.Join(errors.New("invalid token"), unexpected.InternalErr)
}

func (s *rsaJRS) SignToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", errors.Join(err, unexpected.InternalErr)
	}
	return ss, nil
}

func (s *rsaJRS) PublicKey() ([]byte, error) {
	publicKeyPEM := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(s.publicKey),
	}
	return pem.EncodeToMemory(&publicKeyPEM), nil
}

func WithClaims(ctx context.Context, claims SPSAccessClaims) context.Context {
	if c, ok := ctx.(interface{ Set(string, any) }); ok {
		c.Set(ClaimsContextKey, claims)
		return ctx
	}

	return context.WithValue(ctx, ClaimsContextKey, claims)
}

func ClaimsFromContext(ctx context.Context) *SPSAccessClaims {
	claims, ok := ctx.Value(ClaimsContextKey).(SPSAccessClaims)
	if !ok {
		return nil
	}
	return &claims
}
