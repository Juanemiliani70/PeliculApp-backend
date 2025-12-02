package utils

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/database"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	Role      string
	UserId    string
	jwt.RegisteredClaims
}

var SECRET_KEY string = os.Getenv("SECRET_KEY")
var SECRET_REFRESH_KEY string = os.Getenv("SECRET_REFRESH_KEY")

func GenerateAllTokens(email, firstName, lastName, role, userId string) (string, string, error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserId:    userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "PeliculApp",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshClaims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserId:    userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "PeliculApp",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(SECRET_REFRESH_KEY))
	if err != nil {
		return "", "", err
	}

	return signedToken, signedRefreshToken, nil
}

// Actualiza tokens en la base de datos
func UpdateAllTokens(userId, token, refreshToken string, client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	updateAt := time.Now().Format(time.RFC3339)
	updateData := bson.M{
		"$set": bson.M{
			"token":         token,
			"refresh_token": refreshToken,
			"update_at":     updateAt,
		},
	}

	userCollection := database.OpenCollection("users", client)
	_, err := userCollection.UpdateOne(ctx, bson.M{"user_id": userId}, updateData)
	return err
}

// Obtiene el access token desde las cookies
func GetAccessToken(c *gin.Context) (string, error) {
	tokenString, err := c.Cookie("access_token")
	if err != nil {
		return "", errors.New("no se pudo obtener access_token de la cookie")
	}
	return tokenString, nil
}

// Valida un token JWT normal
func ValidateToken(tokenString string) (*SignedDetails, error) {
	claims := &SignedDetails{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return nil, err
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("método de firma inválido")
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("el token ha expirado")
	}

	return claims, nil
}

// Valida refresh token
func ValidateRefreshToken(tokenString string) (*SignedDetails, error) {
	claims := &SignedDetails{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_REFRESH_KEY), nil
	})
	if err != nil {
		return nil, err
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("método de firma inválido")
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("el refresh token ha expirado")
	}

	return claims, nil
}

// Obtener userId desde contexto de gin
func GetUserIdFromContext(c *gin.Context) (string, error) {
	userId, exists := c.Get("userId")
	if !exists {
		return "", errors.New("userId no existe en este contexto")
	}
	id, ok := userId.(string)
	if !ok {
		return "", errors.New("no se puede obtener userId")
	}
	return id, nil
}

// Obtener role desde contexto de gin
func GetRoleFromContext(c *gin.Context) (string, error) {
	role, exists := c.Get("role")
	if !exists {
		return "", errors.New("role no existe en este contexto")
	}
	memberRole, ok := role.(string)
	if !ok {
		return "", errors.New("no se puede obtener role")
	}
	return memberRole, nil
}
