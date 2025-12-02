package routes

import (
	controller "github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/controllers"
	"github.com/Juanemiliani70/PeliculApp/Server/PeliculAppServer/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupProtectedRoutes(router *gin.Engine, client *mongo.Client) {

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleWare())
	protected.POST("/addmovie", controller.AddMovie(client))
	protected.PATCH("/updatereview/:imdb_id", controller.AdminReview(client))
}
