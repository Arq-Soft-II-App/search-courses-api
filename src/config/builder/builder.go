package builder

import (
	"search-courses-api/src/clients"
	"search-courses-api/src/config/envs"
	"search-courses-api/src/config/rabbitmq"
	"search-courses-api/src/controllers"
	"search-courses-api/src/middlewares"
	"search-courses-api/src/routes"
	"search-courses-api/src/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AppBuilder struct {
	envs          envs.Envs
	rabbitMQ      *rabbitmq.RabbitMQ
	solrClient    *clients.SolrClient
	searchService *services.SearchService
	searchCtrl    *controllers.SearchController
	router        *gin.Engine
	logger        *zap.Logger
}

func BuildApp() *AppBuilder {
	builder := &AppBuilder{}
	builder.envs = envs.LoadEnvs(".env")
	builder.BuildLogger()
	builder.BuildRabbitMQ()
	builder.BuildSolrClient()
	builder.BuildServices()
	builder.BuildControllers()
	builder.BuildRouter()
	return builder
}

func (b *AppBuilder) BuildLogger() {
	logger, _ := zap.NewProduction()
	b.logger = logger
}

func (b *AppBuilder) BuildRabbitMQ() {
	b.rabbitMQ = rabbitmq.NewRabbitMQ()
}

func (b *AppBuilder) BuildSolrClient() {
	var err error
	b.solrClient, err = clients.NewSolrClient(b.logger)
	if err != nil {
		b.logger.Fatal("Error al conectar con Solr", zap.Error(err))
	}
}

func (b *AppBuilder) BuildServices() {
	coursesAPIURL := b.envs.Get("COURSES_API_URL")
	if coursesAPIURL == "" {
		coursesAPIURL = "http://localhost:4002"
	}

	b.searchService = services.NewSearchService(b.solrClient, b.logger, coursesAPIURL)
}

func (b *AppBuilder) BuildControllers() {
	b.searchCtrl = controllers.NewSearchController(b.searchService)
}

func (b *AppBuilder) BuildRouter() {
	b.router = gin.Default()

	// Aplicar middlewares aquí
	b.router.Use(middlewares.LoggerMiddleware(b.logger))
	b.router.Use(middlewares.ErrorHandlerMiddleware(b.logger))
	b.router.Use(middlewares.APIKeyAuthMiddleware(b.logger))

	routes.SetupRoutes(b.router, b.searchCtrl)
}

func (b *AppBuilder) GetRabbitMQ() *rabbitmq.RabbitMQ {
	return b.rabbitMQ
}

func (b *AppBuilder) GetSearchService() *services.SearchService {
	return b.searchService
}

func (b *AppBuilder) GetLogger() *zap.Logger {
	return b.logger
}

func (b *AppBuilder) GetRouter() *gin.Engine {
	return b.router
}

func (b *AppBuilder) GetPort() string {
	port := b.envs.Get("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}
