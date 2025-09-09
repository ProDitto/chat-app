package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Custom claims struct for email verification
type EmailVerificationClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// NewVerificationClaims creates new claims for email verification
func NewVerificationClaims(userID, email string, expires time.Time) *EmailVerificationClaims {
	return &EmailVerificationClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func GenerateJWT(userID, username, tokenUUID, secret string, expires time.Time) (string, error) {
	claims := jwt.MapClaims{
		"authorized": true,
		"user_id":    userID,
		"username":   username,
		"exp":        expires.Unix(),
	}

	isRefreshToken := expires.Sub(time.Now()) > time.Hour*24 // Heuristic for refresh token
	if isRefreshToken {
		claims["refresh_uuid"] = tokenUUID
	} else {
		claims["access_uuid"] = tokenUUID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateJWTWithClaims generates a JWT with custom claims.
func GenerateJWTWithClaims(claims jwt.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateEmailVerificationToken specifically validates an email verification token.
func ValidateEmailVerificationToken(tokenString, secret string) (*EmailVerificationClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &EmailVerificationClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*EmailVerificationClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid email verification token")
}
