// Package handlers provides HTTP request handlers
package handlers

import (
	"context"
	"net/http"
	"time"

	"quizmasterapi/config"
	"quizmasterapi/models"
	"quizmasterapi/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// QuizHandler handles quiz-related requests
type QuizHandler struct {
	collection    *mongo.Collection
	courseService *services.CourseService
}

// NewQuizHandler creates a new quiz handler
func NewQuizHandler() *QuizHandler {
	return &QuizHandler{
		collection:    config.GetCollection("quizzes"),
		courseService: services.NewCourseService(),
	}
}

// CreateQuizRequest represents the request to create a quiz
type CreateQuizRequest struct {
	Title           string                  `json:"title" binding:"required" example:"Introduction to Go Programming"`
	Description     string                  `json:"description" example:"Basic concepts of Go programming language"`
	Category        models.QuizCategory     `json:"category" binding:"required" example:"programming"`
	DifficultyLevel models.DifficultyLevel  `json:"difficulty_level" binding:"required" enums:"easy,medium,hard" example:"easy"`
	CourseID        string                  `json:"course_id" binding:"required" example:"course123"`
	Questions       []CreateQuestionRequest `json:"questions" binding:"required,min=1"`
}

// CreateQuestionRequest represents a question in the create quiz request
type CreateQuestionRequest struct {
	QuestionText  string              `json:"question_text" binding:"required" example:"Is Go a statically typed language?"`
	Type          models.QuestionType `json:"type" binding:"required" enums:"true_false,multiple_choice" example:"true_false"`
	Options       []string            `json:"options,omitempty" swaggertype:"array,string"`
	CorrectAnswer string              `json:"correct_answer" binding:"required" example:"true"`
	Points        int                 `json:"points" binding:"required,min=1" example:"10"`
}

// CreateQuiz godoc
// @Summary      Create a new quiz
// @Description  Create a new quiz. Professors create approved quizzes, students create pending quizzes
// @Tags         quizzes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateQuizRequest true "Quiz details"
// @Success      201 {object} models.Quiz
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /quizzes [post]
func (h *QuizHandler) CreateQuiz(c *gin.Context) {
	var req CreateQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	objectID := userID.(primitive.ObjectID)
	role := userRole.(models.UserRole)

	// Validate questions
	questions := make([]models.Question, len(req.Questions))
	for i, q := range req.Questions {
		// Validate question type
		if q.Type != models.QuestionTypeTrueFalse && q.Type != models.QuestionTypeMultipleChoice {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question type"})
			return
		}

		// Validate multiple choice has options
		if q.Type == models.QuestionTypeMultipleChoice && len(q.Options) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Multiple choice questions must have at least 2 options"})
			return
		}

		questions[i] = models.Question{
			ID:            primitive.NewObjectID(),
			QuestionText:  q.QuestionText,
			Type:          q.Type,
			Options:       q.Options,
			CorrectAnswer: q.CorrectAnswer,
			TimeLimit:     15, // Default 15 seconds
			Points:        q.Points,
			Order:         i + 1,
		}
	}

	// Determine quiz status based on creator role
	status := models.StatusApproved
	if role == models.RoleStudent {
		status = models.StatusPending
	}

	quiz := models.Quiz{
		ID:              primitive.NewObjectID(),
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		DifficultyLevel: req.DifficultyLevel,
		CourseID:        req.CourseID,
		CreatorID:       objectID,
		CreatorRole:     role,
		Status:          status,
		Questions:       questions,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := h.collection.InsertOne(ctx, quiz)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create quiz"})
		return
	}

	c.JSON(http.StatusCreated, quiz)
}

// GetQuizzes godoc
// @Summary      List quizzes
// @Description  Get a list of quizzes with optional filters (category, difficulty, status)
// @Tags         quizzes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        category query string false "Filter by category"
// @Param        difficulty query string false "Filter by difficulty level" Enums(easy, medium, hard)
// @Param        status query string false "Filter by status (professors only)" Enums(pending, approved, rejected)
// @Success      200 {array} models.Quiz
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /quizzes [get]
func (h *QuizHandler) GetQuizzes(c *gin.Context) {
	category := c.Query("category")
	difficulty := c.Query("difficulty")
	status := c.Query("status")
	userRole, _ := c.Get("user_role")

	filter := bson.M{}

	// Only show approved quizzes to students
	if userRole.(models.UserRole) == models.RoleStudent {
		filter["status"] = models.StatusApproved
	} else if status != "" {
		filter["status"] = status
	}

	if category != "" {
		filter["category"] = category
	}

	if difficulty != "" {
		filter["difficulty_level"] = difficulty
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := h.collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quizzes"})
		return
	}
	defer cursor.Close(ctx)

	var quizzes []models.Quiz
	if err := cursor.All(ctx, &quizzes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode quizzes"})
		return
	}

	c.JSON(http.StatusOK, quizzes)
}

// GetQuizByID godoc
// @Summary      Get quiz by ID
// @Description  Get detailed information about a specific quiz
// @Tags         quizzes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Quiz ID"
// @Success      200 {object} models.Quiz
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /quizzes/{id} [get]
func (h *QuizHandler) GetQuizByID(c *gin.Context) {
	quizID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(quizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var quiz models.Quiz
	err = h.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&quiz)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}

	userRole, _ := c.Get("user_role")
	// Students can only view approved quizzes
	if userRole.(models.UserRole) == models.RoleStudent && quiz.Status != models.StatusApproved {
		c.JSON(http.StatusForbidden, gin.H{"error": "Quiz not available"})
		return
	}

	c.JSON(http.StatusOK, quiz)
}

// ApproveQuizRequest represents the request to approve/reject a quiz
type ApproveQuizRequest struct {
	Status models.QuizStatus `json:"status" binding:"required" enums:"approved,rejected" example:"approved"`
}

// ApproveQuiz godoc
// @Summary      Approve or reject a quiz
// @Description  Approve or reject a pending quiz (professors only)
// @Tags         quizzes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Quiz ID"
// @Success      200 {object} models.Quiz
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /quizzes/{id}/approve [put]
// ApproveQuiz handles quiz approval/rejection (professors only)
func (h *QuizHandler) ApproveQuiz(c *gin.Context) {
	quizID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(quizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	var req ApproveQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status != models.StatusApproved && req.Status != models.StatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'approved' or 'rejected'"})
		return
	}

	userID, _ := c.Get("user_id")
	approverID := userID.(primitive.ObjectID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":      req.Status,
			"approved_by": approverID,
			"approved_at": now,
			"updated_at":  now,
		},
	}

	result, err := h.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quiz status updated successfully"})
}

// DeleteQuiz godoc
// @Summary      Delete a quiz
// @Description  Delete a quiz (creator or professor only)
// @Tags         quizzes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Quiz ID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /quizzes/{id} [delete]
func (h *QuizHandler) DeleteQuiz(c *gin.Context) {
	quizID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(quizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if quiz exists and user has permission
	var quiz models.Quiz
	err = h.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&quiz)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}

	// Only creator or professors can delete
	if userRole.(models.UserRole) != models.RoleProfessor && quiz.CreatorID != userID.(primitive.ObjectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this quiz"})
		return
	}

	_, err = h.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete quiz"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quiz deleted successfully"})
}
