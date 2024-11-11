package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type SearchCourseModel struct {
	ID                primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	CourseName        string             `json:"course_name" bson:"course_name"`
	CourseDescription string             `json:"description" bson:"description"`
	CoursePrice       float64            `json:"price" bson:"price"`
	CourseDuration    int                `json:"duration" bson:"duration"`
	CourseInitDate    string             `json:"init_date" bson:"init_date"`
	CourseState       bool               `json:"state" bson:"state"`
	CourseCapacity    int                `json:"capacity" bson:"capacity"`
	CourseImage       string             `json:"image" bson:"image"`
	CategoryID        primitive.ObjectID `json:"category_id" bson:"category_id"`
	CategoryName      string             `json:"category_name" bson:"category_name,omitempty"`
	RatingAvg         float64            `json:"ratingavg" bson:"ratingavg,omitempty"`
}
