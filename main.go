package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/database"
	"github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Advertencia: no se encontró el archivo .env")
	}
}

func main() {
	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, PeliculApp!")
	})

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
			log.Println("Allowed Origin:", origins[i])
		}
	} else {
		origins = []string{"http://localhost:5173"}
		log.Println("Allowed Origin: http://localhost:5173")
	}

	config := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(config))
	router.Use(gin.Logger())

	// Conexión a MongoDB
	client := database.Connect()
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatalf("Falló la conexión a MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("No se pudo desconectar de MongoDB: %v", err)
		}
	}()

	// Rutas
	routes.SetupUnProtectedRoutes(router, client)
	routes.SetupProtectedRoutes(router, client)

	// Levantar servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Servidor corriendo en puerto %s", port)
	if err := router.Run(":" + port); err != nil {
		fmt.Println("Falló el inicio del servidor:", err)
	}
}
