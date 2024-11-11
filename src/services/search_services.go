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
	// Obtener datos del curso desde courses-api
	courseData, err := s.getCourseByID(courseID)
	if err != nil {
		s.logger.Error("Error al obtener datos del curso", zap.String("courseID", courseID), zap.Error(err))
		return err
	}

	// Actualizar Solr con los datos del curso
	err = s.solrClient.AddCourse(courseData)
	if err != nil {
		s.logger.Error("Error al actualizar Solr", zap.String("courseID", courseID), zap.Error(err))
		return err
	}

	s.logger.Info("Datos del curso actualizados en Solr", zap.String("courseID", courseID))
	return nil
}

func (s *SearchService) LoadAllCoursesIntoSolr() error {
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

	fmt.Println("course: ", course)

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

	fmt.Println("courses en getAllCourses: ", courses)

	return courses, nil
}

func (s *SearchService) SearchCourses(query string) ([]models.SearchCourseModel, error) {
	courses, err := s.solrClient.SearchCourses(query)
	if err != nil {
		s.logger.Error("Error al buscar cursos en Solr", zap.Error(err))
		return nil, err
	}
	return courses, nil
}
