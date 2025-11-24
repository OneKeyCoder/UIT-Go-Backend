package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents JWT claims
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID int, email, role, secret string, accessExpiry, refreshExpiry time.Duration) (*TokenPair, error) {
	// Generate access token
	accessToken, err := GenerateToken(userID, email, role, secret, accessExpiry)
	if err != nil {
		return nil, err
	}
	
	// Generate refresh token (without role for security)
	refreshToken, err := GenerateToken(userID, email, "", secret, refreshExpiry)
	if err != nil {
		return nil, err
	}
	
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GenerateToken generates a JWT token
func GenerateToken(userID int, email, role, secret string, expiry time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	
	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token
func RefreshAccessToken(refreshToken, secret string, accessExpiry time.Duration) (string, error) {
	claims, err := ValidateToken(refreshToken, secret)
	if err != nil {
		return "", err
	}
	
	// Generate new access token
	return GenerateToken(claims.UserID, claims.Email, claims.Role, secret, accessExpiry)
}
