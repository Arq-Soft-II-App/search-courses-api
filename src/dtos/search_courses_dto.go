package dtos

type SearchCourseDto struct {
	CourseId          string  `json:"_id"`
	CourseName        string  `json:"course_name"`
	CourseDescription string  `json:"description"`
	CoursePrice       float64 `json:"price"`
	CourseDuration    int     `json:"duration"`
	CourseInitDate    string  `json:"init_date"`
	CourseState       bool    `json:"state"`
	CourseCapacity    int     `json:"capacity"`
	CourseImage       string  `json:"image"`
	CategoryID        string  `json:"category_id"`
	CategoryName      string  `json:"category_name"`
	RatingAvg         float64 `json:"ratingavg"`
}

type SearchCourseResponseDto struct {
	Course SearchCourseDto `json:"course"`
}

type SearchCoursesResponseDto struct {
	Courses []SearchCourseDto `json:"courses"`
}
