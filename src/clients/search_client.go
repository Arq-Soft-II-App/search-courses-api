package clients

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"search-courses-api/src/config/envs"
	"search-courses-api/src/models"

	"github.com/vanng822/go-solr/solr"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type SolrClient struct {
	connection *solr.SolrInterface
	logger     *zap.Logger
	mu         sync.RWMutex
	connected  bool
	connCond   *sync.Cond
}

func NewSolrClient(logger *zap.Logger) *SolrClient {
	client := &SolrClient{
		logger: logger,
	}
	client.connCond = sync.NewCond(&client.mu)
	go client.connectWithRetry()
	return client
}

func (s *SolrClient) connectWithRetry() {
	envs := envs.LoadEnvs(".env")
	solrHost := envs.Get("SOLR_HOST")
	solrPort := envs.Get("SOLR_PORT")
	solrCore := envs.Get("SOLR_CORE")

	solrBaseURL := fmt.Sprintf("http://%s:%s/solr", solrHost, solrPort)

	for {
		solrInterface, err := solr.NewSolrInterface(solrBaseURL, solrCore)
		if err != nil {
			s.logger.Error("[SEARCH-API] Error al conectar con Solr", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		s.mu.Lock()
		s.connection = solrInterface
		s.connected = true
		s.mu.Unlock()

		s.logger.Info("[SEARCH-API] Conexión a Solr establecida", zap.String("url", solrBaseURL))

		// Notificar a quienes estén esperando que la conexión está lista
		s.connCond.Broadcast()

		break
	}
}

func (s *SolrClient) WaitForConnection() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for !s.connected {
		s.connCond.Wait()
	}
}

func (s *SolrClient) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

func (s *SolrClient) AddCourse(course *models.SearchCourseModel) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.connected {
		return fmt.Errorf("Conexión a Solr no establecida")
	}

	s.logger.Info("Agregando curso a Solr",
		zap.String("course_id", course.ID.Hex()),
		zap.String("course_name", course.CourseName))

	doc := solr.Document{
		"id":            course.ID.Hex(),
		"course_name":   course.CourseName,
		"description":   course.CourseDescription,
		"price":         course.CoursePrice,
		"duration":      course.CourseDuration,
		"init_date":     course.CourseInitDate,
		"state":         course.CourseState,
		"capacity":      course.CourseCapacity,
		"image":         course.CourseImage,
		"category_id":   course.CategoryID.Hex(),
		"category_name": course.CategoryName,
		"ratingavg":     course.RatingAvg,
	}

	docs := []solr.Document{doc}

	_, err := s.connection.Add(docs, 0, nil)
	if err != nil {
		s.logger.Error("Error al agregar curso a Solr",
			zap.String("course_id", course.ID.Hex()),
			zap.Error(err))
		return err
	}

	s.logger.Info("Curso agregado exitosamente a Solr",
		zap.String("course_id", course.ID.Hex()))

	// Commit los cambios
	_, err = s.connection.Commit()
	if err != nil {
		s.logger.Error("Error al hacer commit en Solr", zap.Error(err))
		return err
	}

	return nil
}

func (s *SolrClient) SearchCourses(query string) ([]models.SearchCourseModel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.connected {
		return nil, fmt.Errorf("Conexión a Solr no establecida")
	}

	// Crear una nueva query de Solr
	solrQuery := solr.NewQuery()

	// Construir la query en base al parámetro recibido
	if query != "" {
		// Escapar caracteres especiales en la query
		escapedQuery := strings.Replace(query, ":", "\\:", -1)
		escapedQuery = strings.Replace(escapedQuery, " ", "\\ ", -1)

		// Búsqueda en múltiples campos
		solrQuery.Q(fmt.Sprintf("(course_name:*%s* OR description:*%s* OR category_name:*%s*)", escapedQuery, escapedQuery, escapedQuery))
	} else {
		// Si no se envía query, devolver todos los resultados
		solrQuery.Q("*:*")
	}

	// Limitar el número de resultados
	solrQuery.Rows(100)

	// Log de la consulta que se ejecutará en Solr
	s.logger.Info("[SEARCH-API] Ejecutando búsqueda en Solr",
		zap.String("query", solrQuery.String()))

	// Ejecutar la búsqueda
	response := s.connection.Search(solrQuery)
	res, err := response.Result(nil)
	if err != nil {
		// Log de error en caso de fallo
		s.logger.Error("[SEARCH-API] Error al ejecutar búsqueda en Solr",
			zap.String("query", query),
			zap.Error(err))
		return nil, err
	}

	// Log de la respuesta recibida
	s.logger.Debug("[SEARCH-API] Respuesta recibida de Solr",
		zap.Int("total_resultados", len(res.Results.Docs)))

	// Procesar los documentos obtenidos y convertirlos en el modelo de la aplicación
	var courses []models.SearchCourseModel
	if res != nil && res.Results != nil {
		for _, doc := range res.Results.Docs {
			s.logger.Info("Procesando documento Solr", zap.Any("doc", doc))

			course := models.SearchCourseModel{}

			// Parsear ID del curso
			if idStr, ok := doc["id"].(string); ok {
				if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
					course.ID = oid
				}
			}

			// Asignar valores de los campos
			course.CourseName = getStringValue(doc, "course_name")
			course.CourseDescription = getStringValue(doc, "description")
			course.CoursePrice = getFloat64Value(doc, "price")
			course.CourseDuration = getIntValue(doc, "duration")
			course.CourseInitDate = getStringValue(doc, "init_date")
			course.CourseState = getBoolValue(doc, "state")
			course.CourseCapacity = getIntValue(doc, "capacity")
			course.CourseImage = getStringValue(doc, "image")

			// Parsear categoría del curso
			if categoryIDStr := getStringValue(doc, "category_id"); categoryIDStr != "" {
				if categoryID, err := primitive.ObjectIDFromHex(categoryIDStr); err == nil {
					course.CategoryID = categoryID
				}
			}

			course.CategoryName = getStringValue(doc, "category_name")
			course.RatingAvg = getFloat64Value(doc, "ratingavg")

			// Agregar el curso a la lista
			courses = append(courses, course)
		}
	}

	// Retornar los cursos encontrados
	return courses, nil
}

func getStringValue(doc map[string]interface{}, key string) string {
	if val, ok := doc[key].([]interface{}); ok && len(val) > 0 {
		if strVal, ok := val[0].(string); ok {
			return strVal
		}
	} else if val, ok := doc[key].(string); ok {
		return val
	}
	return ""
}

func getFloat64Value(doc map[string]interface{}, key string) float64 {
	if val, ok := doc[key].([]interface{}); ok && len(val) > 0 {
		if numVal, ok := val[0].(float64); ok {
			return numVal
		}
	} else if val, ok := doc[key].(float64); ok {
		return val
	}
	return 0
}

func getIntValue(doc map[string]interface{}, key string) int {
	if val, ok := doc[key].([]interface{}); ok && len(val) > 0 {
		if numVal, ok := val[0].(float64); ok {
			return int(numVal)
		}
	} else if val, ok := doc[key].(float64); ok {
		return int(val)
	}
	return 0
}

func getBoolValue(doc map[string]interface{}, key string) bool {
	if val, ok := doc[key].([]interface{}); ok && len(val) > 0 {
		if boolVal, ok := val[0].(bool); ok {
			return boolVal
		}
	} else if val, ok := doc[key].(bool); ok {
		return val
	}
	return false
}
