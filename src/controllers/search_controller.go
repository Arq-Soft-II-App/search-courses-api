package controllers

import (
	"fmt"
	"net/http"
	"search-courses-api/src/dtos"
	"search-courses-api/src/services"

	"github.com/gin-gonic/gin"
)

type SearchController struct {
	searchService *services.SearchService
}

func NewSearchController(searchService *services.SearchService) *SearchController {
	return &SearchController{
		searchService: searchService,
	}
}

func (s *SearchController) SearchCourses(c *gin.Context) {
	query := c.Query("q")
	courses, err := s.searchService.SearchCourses(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar cursos"})
		return
	}

	fmt.Println("courses: ", courses)

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

	c.JSON(http.StatusOK, responseDto)
}
