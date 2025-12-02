package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/database"
	"github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/models"
	"github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var validate = validator.New()

// Obtener todas las películas
func GetMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		movieCollection := database.OpenCollection("movies", client)

		cursor, err := movieCollection.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener las películas."})
			return
		}
		defer cursor.Close(ctx)

		var movies []models.Movie
		if err := cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al decodificar las películas."})
			return
		}

		c.JSON(http.StatusOK, movies)
	}
}

// Obtener una película por IMDB ID
func GetMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		movieID := c.Param("imdb_id")
		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere el ID de la película"})
			return
		}

		movieCollection := database.OpenCollection("movies", client)
		var movie models.Movie

		err := movieCollection.FindOne(ctx, bson.D{{Key: "imdb_id", Value: movieID}}).Decode(&movie)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Película no encontrada"})
			return
		}

		c.JSON(http.StatusOK, movie)
	}
}

// Buscar películas por título o género
func SearchMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		query := strings.TrimSpace(c.Query("title"))
		if query == "" {
			query = strings.TrimSpace(c.Query("query"))
		}

		genre := strings.TrimSpace(c.Query("genre"))

		movieCollection := database.OpenCollection("movies", client)

		filter := bson.M{}

		if query != "" {
			filter["title"] = bson.M{
				"$regex":   query,
				"$options": "i",
			}
		}

		if genre != "" {
			filter["genre.genre_name"] = bson.M{
				"$regex":   genre,
				"$options": "i",
			}
		}

		findOptions := options.Find()
		findOptions.SetCollation(&options.Collation{
			Locale:   "es",
			Strength: 1,
		})

		cursor, err := movieCollection.Find(ctx, filter, findOptions)
		if err != nil {
			c.JSON(http.StatusOK, []models.Movie{})
			return
		}
		defer cursor.Close(ctx)

		var movies []models.Movie
		if err := cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusOK, []models.Movie{})
			return
		}

		c.JSON(http.StatusOK, movies)
	}
}

// Agregar una película
func AddMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var movie models.Movie
		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de entrada inválidos"})
			return
		}

		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validación fallida", "detalles": err.Error()})
			return
		}

		movieCollection := database.OpenCollection("movies", client)
		result, err := movieCollection.InsertOne(ctx, movie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo agregar la película"})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

func AdminReview(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := utils.GetRoleFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No se encontró el rol en el contexto"})
			return
		}

		if role != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "El usuario debe ser ADMIN"})
			return
		}

		movieID := c.Param("imdb_id")
		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere el ID de la película"})
			return
		}

		var req struct {
			AdminReview string `json:"admin_review"`
		}
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de solicitud inválido"})
			return
		}

		filter := bson.D{{Key: "imdb_id", Value: movieID}}
		update := bson.M{
			"$set": bson.M{
				"admin_review": req.AdminReview, // solo guardamos la reseña
			},
		}

		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		movieCollection := database.OpenCollection("movies", client)
		result, err := movieCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar la película"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Película no encontrada"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"admin_review": req.AdminReview})
	}
}

// Obtener géneros
func GetGenres(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		genreCollection := database.OpenCollection("genres", client)
		cursor, err := genreCollection.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener géneros"})
			return
		}
		defer cursor.Close(ctx)

		var genres []models.Genre
		if err := cursor.All(ctx, &genres); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, genres)
	}
}
