package routes

import (
	"net/http"
	"search-courses-api/src/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, searchController *controllers.SearchController) {
	searchRoutes := router.Group("/search")
	{
		searchRoutes.GET("/", searchController.SearchCourses)
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada"})
	})
}
