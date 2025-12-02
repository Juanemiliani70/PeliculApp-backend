package routes

import (
	controller "github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/controllers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupUnProtectedRoutes(router *gin.Engine, client *mongo.Client) {

	router.GET("/movies", controller.GetMovies(client))
	router.GET("/movie/:imdb_id", controller.GetMovie(client))
	router.GET("/genres", controller.GetGenres(client))
	router.GET("/search", controller.SearchMovies(client))
	router.POST("/register", controller.RegisterUser(client))
	router.POST("/login", controller.LoginUser(client))
	router.POST("/logout", controller.LogoutHandler(client))
	router.POST("/refresh", controller.RefreshTokenHandler(client))
}
