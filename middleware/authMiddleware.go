package middleware

import (
	"log"
	"net/http"

	"github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener token de acceso desde la cookie
		token, err := utils.GetAccessToken(c)
		log.Println("Auth token recibido:", token) // debug del token
		if err != nil {
			log.Println("Error al obtener token:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Error al obtener token: " + err.Error()})
			c.Abort()
			return
		}

		if token == "" {
			log.Println("Token vacío recibido")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No se proporcionó token"})
			c.Abort()
			return
		}

		// Validar token
		claims, err := utils.ValidateToken(token)
		if err != nil {
			log.Println("Token inválido:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido: " + err.Error()})
			c.Abort()
			return
		}

		log.Println("Token válido para userId:", claims.UserId, "role:", claims.Role)

		// Guardar información del usuario en el contexto
		c.Set("userId", claims.UserId)
		c.Set("role", claims.Role)

		c.Next()
	}
}
