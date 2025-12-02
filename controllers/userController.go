package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/database"
	"github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/models"
	"github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func RegisterUser(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de entrada inválidos"})
			return
		}

		if err := validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validación fallida", "detalles": err.Error()})
			return
		}

		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo encriptar la contraseña"})
			return
		}

		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		userCollection := database.OpenCollection("users", client)

		count, err := userCollection.CountDocuments(ctx, bson.D{{Key: "email", Value: user.Email}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar usuario existente"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "El usuario ya existe"})
			return
		}

		user.UserID = bson.NewObjectID().Hex()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		user.Password = hashedPassword

		result, err := userCollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear el usuario"})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

func LoginUser(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userLogin models.UserLogin
		if err := c.ShouldBindJSON(&userLogin); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de entrada inválidos"})
			return
		}

		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		userCollection := database.OpenCollection("users", client)

		var foundUser models.User
		err := userCollection.FindOne(ctx, bson.D{{Key: "email", Value: userLogin.Email}}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Correo o contraseña inválidos"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(userLogin.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Correo o contraseña inválidos"})
			return
		}

		token, refreshToken, err := utils.GenerateAllTokens(foundUser.Email, foundUser.FirstName, foundUser.LastName, foundUser.Role, foundUser.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar tokens"})
			return
		}

		err = utils.UpdateAllTokens(foundUser.UserID, token, refreshToken, client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar tokens"})
			return
		}

		// Cookies con SameSite=Lax para localhost
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "access_token",
			Value:    token,
			Path:     "/",
			MaxAge:   86400,
			Secure:   false,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/",
			MaxAge:   604800,
			Secure:   false,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		c.JSON(http.StatusOK, models.UserResponse{
			UserID:          foundUser.UserID,
			FirstName:       foundUser.FirstName,
			LastName:        foundUser.LastName,
			Email:           foundUser.Email,
			Role:            foundUser.Role,
			FavouriteGenres: foundUser.FavouriteGenres,
		})
	}
}

func LogoutHandler(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var logoutRequest struct {
			UserId string `json:"user_id"`
		}
		if err := c.ShouldBindJSON(&logoutRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
			return
		}

		err := utils.UpdateAllTokens(logoutRequest.UserId, "", "", client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al cerrar sesión"})
			return
		}

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "access_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Secure:   false,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Secure:   false,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		c.JSON(http.StatusOK, gin.H{"message": "Sesión cerrada correctamente"})
	}
}

func RefreshTokenHandler(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No se pudo obtener el token de actualización"})
			return
		}

		claim, err := utils.ValidateRefreshToken(refreshToken)
		if err != nil || claim == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token de actualización inválido o expirado"})
			return
		}

		userCollection := database.OpenCollection("users", client)
		var user models.User
		err = userCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: claim.UserId}}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
			return
		}

		newToken, newRefreshToken, _ := utils.GenerateAllTokens(user.Email, user.FirstName, user.LastName, user.Role, user.UserID)
		err = utils.UpdateAllTokens(user.UserID, newToken, newRefreshToken, client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar tokens"})
			return
		}

		c.SetCookie("access_token", newToken, 86400, "/", "localhost", false, true)
		c.SetCookie("refresh_token", newRefreshToken, 604800, "/", "localhost", false, true)

		c.JSON(http.StatusOK, gin.H{"message": "Tokens actualizados correctamente"})
	}
}
