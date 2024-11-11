package clients

import (
	"fmt"
	"strings"
	"sync"

	"search-courses-api/src/config/envs"
	"search-courses-api/src/models"

	"github.com/vanng822/go-solr/solr"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type SolrClient struct {
	connection *solr.SolrInterface
	logger     *zap.Logger
}

var (
	solrClientInstance *SolrClient
	solrOnce           sync.Once
)

func NewSolrClient(logger *zap.Logger) (*SolrClient, error) {
	var err error
	solrOnce.Do(func() {
		envs := envs.LoadEnvs(".env")
		solrHost := envs.Get("SOLR_HOST")
		solrPort := envs.Get("SOLR_PORT")
		solrCore := envs.Get("SOLR_CORE")

		solrBaseURL := fmt.Sprintf("http://%s:%s/solr", solrHost, solrPort)

		solrInterface, err := solr.NewSolrInterface(solrBaseURL, solrCore)
		if err != nil {
			logger.Error("[SEARCH-API] Error al conectar con Solr", zap.Error(err))
			return
		}

		solrClientInstance = &SolrClient{
			connection: solrInterface,
			logger:     logger,
		}

		logger.Info("[SEARCH-API] Conexión a Solr establecida", zap.String("url", solrBaseURL))
	})

	return solrClientInstance, err
}

func (s *SolrClient) AddCourse(course *models.SearchCourseModel) error {

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
		s.logger.Error("Error al agregar documento a Solr", zap.Error(err))
		return err
	}

	// Commit los cambios
	_, err = s.connection.Commit()
	if err != nil {
		s.logger.Error("Error al hacer commit en Solr", zap.Error(err))
		return err
	}

	return nil
}

func (s *SolrClient) SearchCourses(query string) ([]models.SearchCourseModel, error) {
	solrQuery := solr.NewQuery()
	if query != "" {
		escapedQuery := strings.Replace(query, ":", "\\:", -1)
		escapedQuery = strings.Replace(escapedQuery, " ", "\\ ", -1)
		solrQuery.Q(fmt.Sprintf("course_name:*%s* OR description:*%s* OR category_name:*%s*", escapedQuery, escapedQuery, escapedQuery))
	} else {
		solrQuery.Q("*:*")
	}

	solrQuery.Rows(100)

	s.logger.Info("Solr query", zap.String("query", solrQuery.String()))

	response := s.connection.Search(solrQuery)
	res, err := response.Result(nil)
	if err != nil {
		s.logger.Error("Error al obtener resultados de Solr", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Solr response", zap.Any("response", res))

	var courses []models.SearchCourseModel
	if res != nil && res.Results != nil {
		for _, doc := range res.Results.Docs {
			s.logger.Info("Procesando documento Solr", zap.Any("doc", doc))

			course := models.SearchCourseModel{}

			if idStr, ok := doc["id"].(string); ok {
				if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
					course.ID = oid
				}
			}

			course.CourseName = getStringValue(doc, "course_name")
			course.CourseDescription = getStringValue(doc, "description")
			course.CoursePrice = getFloat64Value(doc, "price")
			course.CourseDuration = getIntValue(doc, "duration")
			course.CourseInitDate = getStringValue(doc, "init_date")
			course.CourseState = getBoolValue(doc, "state")
			course.CourseCapacity = getIntValue(doc, "capacity")
			course.CourseImage = getStringValue(doc, "image")

			if categoryIDStr := getStringValue(doc, "category_id"); categoryIDStr != "" {
				if categoryID, err := primitive.ObjectIDFromHex(categoryIDStr); err == nil {
					course.CategoryID = categoryID
				}
			}

			course.CategoryName = getStringValue(doc, "category_name")
			course.RatingAvg = getFloat64Value(doc, "ratingavg")

			courses = append(courses, course)
		}
	}

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
