// Package models defines the data structures for the quiz application
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	RoleProfessor UserRole = "professor"
	RoleStudent   UserRole = "student"
)

// User represents a user in the system
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	Email     string             `bson:"email" json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password  string             `bson:"password" json:"-"`
	FirstName string             `bson:"first_name" json:"first_name" binding:"required" example:"John"`
	LastName  string             `bson:"last_name" json:"last_name" binding:"required" example:"Doe"`
	Role      UserRole           `bson:"role" json:"role" binding:"required" enums:"student,professor" example:"student"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// QuizCategory represents different categories of quizzes
type QuizCategory string

const (
	CategoryMathematics QuizCategory = "mathematics"
	CategoryScience     QuizCategory = "science"
	CategoryHistory     QuizCategory = "history"
	CategoryLiterature  QuizCategory = "literature"
	CategoryProgramming QuizCategory = "programming"
	CategoryLanguages   QuizCategory = "languages"
)

// DifficultyLevel represents the difficulty level of a quiz
type DifficultyLevel string

const (
	LevelEasy   DifficultyLevel = "easy"
	LevelMedium DifficultyLevel = "medium"
	LevelHard   DifficultyLevel = "hard"
)

// QuizStatus represents the approval status of a quiz
type QuizStatus string

const (
	StatusPending  QuizStatus = "pending"
	StatusApproved QuizStatus = "approved"
	StatusRejected QuizStatus = "rejected"
)

// QuestionType represents the type of question
type QuestionType string

const (
	QuestionTypeTrueFalse      QuestionType = "true_false"
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
)

// Quiz represents a quiz in the system
type Quiz struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title           string             `bson:"title" json:"title" binding:"required"`
	Description     string             `bson:"description" json:"description"`
	Category        QuizCategory       `bson:"category" json:"category" binding:"required"`
	DifficultyLevel DifficultyLevel    `bson:"difficulty_level" json:"difficulty_level" binding:"required"`
	CourseID        string             `bson:"course_id" json:"course_id" binding:"required"` // External course reference
	CreatorID       primitive.ObjectID `bson:"creator_id" json:"creator_id"`
	CreatorRole     UserRole           `bson:"creator_role" json:"creator_role"`
	Status          QuizStatus         `bson:"status" json:"status"`
	Questions       []Question         `bson:"questions" json:"questions"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
	ApprovedBy      primitive.ObjectID `bson:"approved_by,omitempty" json:"approved_by,omitempty"`
	ApprovedAt      *time.Time         `bson:"approved_at,omitempty" json:"approved_at,omitempty"`
}

// Question represents a question in a quiz
type Question struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439012"`
	QuestionText  string             `bson:"question_text" json:"question_text" binding:"required" example:"Is Go a compiled language?"`
	Type          QuestionType       `bson:"type" json:"type" binding:"required" enums:"true_false,multiple_choice" example:"true_false"`
	Options       []string           `bson:"options,omitempty" json:"options,omitempty" swaggertype:"array,string"`    // For multiple choice
	CorrectAnswer interface{}        `bson:"correct_answer" json:"correct_answer" swaggertype:"string" example:"true"` // bool for T/F, int for MC (index)
	TimeLimit     int                `bson:"time_limit" json:"time_limit" example:"15"`                                // In seconds, default 15
	Points        int                `bson:"points" json:"points" example:"10"`                                        // Base points for this question
	Order         int                `bson:"order" json:"order" example:"1"`                                           // Question order in quiz
}

// QuizAttempt represents a student's attempt at a quiz
type QuizAttempt struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	QuizID      primitive.ObjectID `bson:"quiz_id" json:"quiz_id"`
	StudentID   primitive.ObjectID `bson:"student_id" json:"student_id"`
	Answers     []Answer           `bson:"answers" json:"answers"`
	TotalScore  float64            `bson:"total_score" json:"total_score"`
	MaxScore    float64            `bson:"max_score" json:"max_score"`
	StartedAt   time.Time          `bson:"started_at" json:"started_at"`
	CompletedAt *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	TimeTaken   int                `bson:"time_taken" json:"time_taken"` // In seconds
}

// Answer represents a student's answer to a question
type Answer struct {
	QuestionID    primitive.ObjectID `bson:"question_id" json:"question_id"`
	StudentAnswer interface{}        `bson:"student_answer" json:"student_answer"`
	IsCorrect     bool               `bson:"is_correct" json:"is_correct"`
	TimeToAnswer  int                `bson:"time_to_answer" json:"time_to_answer"` // In seconds
	PointsEarned  float64            `bson:"points_earned" json:"points_earned"`
	AnsweredAt    time.Time          `bson:"answered_at" json:"answered_at"`
}

// LeaderboardEntry represents an entry in the quiz leaderboard
type LeaderboardEntry struct {
	Rank        int                `json:"rank"`
	StudentID   primitive.ObjectID `json:"student_id"`
	StudentName string             `json:"student_name"`
	Score       float64            `json:"score"`
	MaxScore    float64            `json:"max_score"`
	Percentage  float64            `json:"percentage"`
	TimeTaken   int                `json:"time_taken"`
	CompletedAt time.Time          `json:"completed_at"`
}
