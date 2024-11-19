package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"search-courses-api/src/clients"
	"search-courses-api/src/models"

	"go.uber.org/zap"
)

type SearchService struct {
	solrClient *clients.SolrClient
	logger     *zap.Logger
	coursesAPI string
}

func NewSearchService(solrClient *clients.SolrClient, logger *zap.Logger, coursesAPI string) *SearchService {
	return &SearchService{
		solrClient: solrClient,
		logger:     logger,
		coursesAPI: coursesAPI,
	}
}

func (s *SearchService) UpdateCourseInSolr(courseID string) error {
	s.logger.Info("[SEARCH-API] Iniciando actualización de curso en Solr",
		zap.String("course_id", courseID))

	if !s.solrClient.IsConnected() {
		return fmt.Errorf("Conexión a Solr no establecida")
	}

	courseData, err := s.getCourseByID(courseID)
	if err != nil {
		s.logger.Error("[SEARCH-API] Error al obtener datos del curso",
			zap.String("course_id", courseID),
			zap.Error(err))
		return err
	}

	s.logger.Debug("[SEARCH-API] Datos del curso obtenidos correctamente",
		zap.String("course_id", courseID),
		zap.String("course_name", courseData.CourseName))

	err = s.solrClient.AddCourse(courseData)
	if err != nil {
		s.logger.Error("Error al actualizar curso en Solr",
			zap.String("course_id", courseID),
			zap.Error(err))
		return err
	}

	s.logger.Info("Curso actualizado exitosamente en Solr",
		zap.String("course_id", courseID))
	return nil
}

func (s *SearchService) LoadAllCoursesIntoSolr() error {
	s.logger.Info("Cargando todos los cursos en Solr")

	courses, err := s.getAllCourses()
	if err != nil {
		s.logger.Error("Error al obtener todos los cursos", zap.Error(err))
		return err
	}

	for _, course := range courses {
		err = s.solrClient.AddCourse(&course)
		if err != nil {
			s.logger.Error("Error al agregar curso a Solr", zap.String("courseID", course.ID.Hex()), zap.Error(err))
		}
	}

	s.logger.Info("Todos los cursos cargados en Solr")
	return nil
}

// Métodos auxiliares para obtener datos desde courses-api

func (s *SearchService) getCourseByID(courseID string) (*models.SearchCourseModel, error) {
	url := fmt.Sprintf("%s/%s", s.coursesAPI, courseID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error al obtener el curso, código de estado: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var course models.SearchCourseModel
	if err := json.Unmarshal(body, &course); err != nil {
		return nil, err
	}

	return &course, nil
}

func (s *SearchService) getAllCourses() ([]models.SearchCourseModel, error) {
	url := fmt.Sprintf("%s/", s.coursesAPI)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error al realizar la solicitud GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error al obtener los cursos, código de estado: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer el cuerpo de la respuesta: %v", err)
	}

	var courses []models.SearchCourseModel
	if err := json.Unmarshal(body, &courses); err != nil {
		return nil, fmt.Errorf("error al deserializar la respuesta JSON: %v", err)
	}

	return courses, nil
}

func (s *SearchService) SearchCourses(query string) ([]models.SearchCourseModel, error) {
	if !s.solrClient.IsConnected() {
		return nil, fmt.Errorf("Servicio de búsqueda no disponible temporalmente")
	}

	s.logger.Info("Iniciando búsqueda de cursos",
		zap.String("query", query))

	courses, err := s.solrClient.SearchCourses(query)
	if err != nil {
		s.logger.Error("Error al buscar cursos",
			zap.String("query", query),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("Búsqueda completada exitosamente",
		zap.String("query", query),
		zap.Int("resultados", len(courses)))
	return courses, nil
}
