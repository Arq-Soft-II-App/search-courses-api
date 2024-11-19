package main

import (
	"log"
	"search-courses-api/src/config/builder"

	"go.uber.org/zap"
)

func main() {
	app := builder.BuildApp()
	logger := app.GetLogger()
	searchService := app.GetSearchService()

	// Iniciar el consumo de mensajes de RabbitMQ
	rabbitMQ := app.GetRabbitMQ()
	rabbitMQ.ConsumeMessages(func(message string) {
		// Procesar cada mensaje (ID de curso)
		err := searchService.UpdateCourseInSolr(message)
		if err != nil {
			logger.Error("Error al actualizar el curso en Solr", zap.Error(err))
		}
	})

	// Esperar a que la conexión con Solr esté lista
	app.GetSolrClient().WaitForConnection()

	// Cargar todos los cursos en Solr al iniciar la aplicación
	err := searchService.LoadAllCoursesIntoSolr()
	if err != nil {
		logger.Error("Error al cargar los cursos en Solr", zap.Error(err))
	}

	// Iniciar el servidor HTTP
	router := app.GetRouter()
	port := app.GetPort()
	if err := router.Run(port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
