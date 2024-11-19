package controllers

import (
	"net/http"
	"search-courses-api/src/dtos"
	"search-courses-api/src/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SearchController struct {
	searchService *services.SearchService
	logger        *zap.Logger
}

func NewSearchController(searchService *services.SearchService, logger *zap.Logger) *SearchController {
	return &SearchController{
		searchService: searchService,
		logger:        logger,
	}
}

func (s *SearchController) SearchCourses(c *gin.Context) {
	query := c.Query("q")
	s.logger.Info("[SEARCH-API] Nueva solicitud de búsqueda recibida",
		zap.String("query", query))

	courses, err := s.searchService.SearchCourses(query)
	if err != nil {
		s.logger.Error("[SEARCH-API] Error al procesar la búsqueda",
			zap.String("query", query),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar cursos"})
		return
	}

	s.logger.Debug("[SEARCH-API] Transformando resultados a DTO",
		zap.Int("total_cursos", len(courses)))

	// Inicializar el slice con capacidad predefinida
	coursesDto := make([]dtos.SearchCourseDto, 0, len(courses))

	for _, course := range courses {
		// Verificar que el ID del curso sea válido
		if course.ID.IsZero() {
			continue // Saltear cursos con ID inválido
		}

		courseDto := dtos.SearchCourseDto{
			CourseId:          course.ID.Hex(),
			CourseName:        course.CourseName,
			CourseDescription: course.CourseDescription,
			CoursePrice:       course.CoursePrice,
			CourseDuration:    course.CourseDuration,
			CourseInitDate:    course.CourseInitDate,
			CourseState:       course.CourseState,
			CourseCapacity:    course.CourseCapacity,
			CourseImage:       course.CourseImage,
			CategoryID:        course.CategoryID.Hex(),
			CategoryName:      course.CategoryName,
			RatingAvg:         course.RatingAvg,
		}
		coursesDto = append(coursesDto, courseDto)
	}

	responseDto := dtos.SearchCoursesResponseDto{
		Courses: coursesDto,
	}

	s.logger.Info("[SEARCH-API] Búsqueda completada exitosamente",
		zap.String("query", query),
		zap.Int("resultados", len(coursesDto)))

	c.JSON(http.StatusOK, responseDto)
}
