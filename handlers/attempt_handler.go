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

// AttemptHandler handles quiz attempt-related requests
type AttemptHandler struct {
	collection     *mongo.Collection
	quizCollection *mongo.Collection
	userCollection *mongo.Collection
	courseService  *services.CourseService
	scoringService *services.ScoringService
}

// NewAttemptHandler creates a new attempt handler
func NewAttemptHandler() *AttemptHandler {
	return &AttemptHandler{
		collection:     config.GetCollection("attempts"),
		quizCollection: config.GetCollection("quizzes"),
		userCollection: config.GetCollection("users"),
		courseService:  services.NewCourseService(),
		scoringService: services.NewScoringService(),
	}
}

// StartAttemptRequest represents the request to start a quiz attempt
type StartAttemptRequest struct {
	QuizID string `json:"quiz_id" binding:"required" example:"507f1f77bcf86cd799439011"`
}

// StartAttempt godoc
// @Summary      Start a quiz attempt
// @Description  Start attempting a quiz (students only, requires course completion)
// @Tags         attempts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body StartAttemptRequest true "Quiz ID to attempt"
// @Success      201 {object} models.QuizAttempt
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      409 {object} map[string]string
// @Router       /attempts/start [post]
func (h *AttemptHandler) StartAttempt(c *gin.Context) {
	var req StartAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quizID, err := primitive.ObjectIDFromHex(req.QuizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	userID, _ := c.Get("user_id")
	studentID := userID.(primitive.ObjectID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get quiz details
	var quiz models.Quiz
	err = h.quizCollection.FindOne(ctx, bson.M{"_id": quizID}).Decode(&quiz)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}

	// Check if quiz is approved
	if quiz.Status != models.StatusApproved {
		c.JSON(http.StatusForbidden, gin.H{"error": "Quiz is not available for attempts"})
		return
	}

	// Check course completion
	completed, err := h.courseService.CheckCourseCompletion(studentID.Hex(), quiz.CourseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify course completion"})
		return
	}

	if !completed {
		c.JSON(http.StatusForbidden, gin.H{"error": "You must complete the required course before attempting this quiz"})
		return
	}

	// Check if student already has an ongoing attempt
	var existingAttempt models.QuizAttempt
	err = h.collection.FindOne(ctx, bson.M{
		"quiz_id":      quizID,
		"student_id":   studentID,
		"completed_at": bson.M{"$exists": false},
	}).Decode(&existingAttempt)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You already have an ongoing attempt for this quiz"})
		return
	}

	// Calculate max score
	maxScore := 0.0
	for _, q := range quiz.Questions {
		maxScore += float64(q.Points)
	}

	// Create new attempt
	attempt := models.QuizAttempt{
		ID:        primitive.NewObjectID(),
		QuizID:    quizID,
		StudentID: studentID,
		Answers:   []models.Answer{},
		MaxScore:  maxScore,
		StartedAt: time.Now(),
	}

	_, err = h.collection.InsertOne(ctx, attempt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start quiz attempt"})
		return
	}

	// Return attempt info with questions (but not correct answers)
	quizForAttempt := quiz
	for i := range quizForAttempt.Questions {
		quizForAttempt.Questions[i].CorrectAnswer = nil
	}

	c.JSON(http.StatusCreated, gin.H{
		"attempt": attempt,
		"quiz":    quizForAttempt,
	})
}

// SubmitAnswerRequest represents a single answer submission
type SubmitAnswerRequest struct {
	AttemptID    string `json:"attempt_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	QuestionID   string `json:"question_id" binding:"required" example:"507f1f77bcf86cd799439012"`
	Answer       string `json:"answer" binding:"required" example:"true"`
	TimeToAnswer int    `json:"time_to_answer" binding:"required" example:"8"` // In seconds
}

// SubmitAnswer godoc
// @Summary      Submit an answer
// @Description  Submit an answer for a specific question in an ongoing attempt
// @Tags         attempts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body SubmitAnswerRequest true "Answer submission"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /attempts/answer [post]
func (h *AttemptHandler) SubmitAnswer(c *gin.Context) {
	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attemptID, err := primitive.ObjectIDFromHex(req.AttemptID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attempt ID"})
		return
	}

	questionID, err := primitive.ObjectIDFromHex(req.QuestionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	userID, _ := c.Get("user_id")
	studentID := userID.(primitive.ObjectID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get attempt
	var attempt models.QuizAttempt
	err = h.collection.FindOne(ctx, bson.M{
		"_id":        attemptID,
		"student_id": studentID,
	}).Decode(&attempt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	// Check if already completed
	if attempt.CompletedAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This attempt is already completed"})
		return
	}

	// Get quiz and question
	var quiz models.Quiz
	err = h.quizCollection.FindOne(ctx, bson.M{"_id": attempt.QuizID}).Decode(&quiz)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}

	var question *models.Question
	for i := range quiz.Questions {
		if quiz.Questions[i].ID == questionID {
			question = &quiz.Questions[i]
			break
		}
	}

	if question == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found in quiz"})
		return
	}

	// Check if answer already submitted
	for _, ans := range attempt.Answers {
		if ans.QuestionID == questionID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Answer already submitted for this question"})
			return
		}
	}

	// Validate time limit
	if req.TimeToAnswer > question.TimeLimit {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Time limit exceeded"})
		return
	}

	// Check if answer is correct
	// Convert both answers to string for comparison
	isCorrect := false
	correctAnswerStr := ""

	switch v := question.CorrectAnswer.(type) {
	case string:
		correctAnswerStr = v
	case bool:
		if v {
			correctAnswerStr = "true"
		} else {
			correctAnswerStr = "false"
		}
	case int:
		correctAnswerStr = string(rune(v + '0'))
	case float64:
		correctAnswerStr = string(rune(int(v) + '0'))
	}

	isCorrect = req.Answer == correctAnswerStr

	// Calculate score
	pointsEarned := h.scoringService.CalculateScore(question.Points, req.TimeToAnswer, isCorrect)

	// Create answer
	answer := models.Answer{
		QuestionID:    questionID,
		StudentAnswer: req.Answer,
		IsCorrect:     isCorrect,
		TimeToAnswer:  req.TimeToAnswer,
		PointsEarned:  pointsEarned,
		AnsweredAt:    time.Now(),
	}

	// Update attempt with new answer
	update := bson.M{
		"$push": bson.M{"answers": answer},
		"$inc":  bson.M{"total_score": pointsEarned},
	}

	_, err = h.collection.UpdateOne(ctx, bson.M{"_id": attemptID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save answer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_correct":    isCorrect,
		"points_earned": pointsEarned,
		"message":       "Answer submitted successfully",
	})
}

// CompleteAttempt godoc
// @Summary      Complete an attempt
// @Description  Mark a quiz attempt as complete and calculate final score
// @Tags         attempts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Attempt ID"
// @Success      200 {object} models.QuizAttempt
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /attempts/{id}/complete [put]
func (h *AttemptHandler) CompleteAttempt(c *gin.Context) {
	attemptID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(attemptID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attempt ID"})
		return
	}

	userID, _ := c.Get("user_id")
	studentID := userID.(primitive.ObjectID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get attempt
	var attempt models.QuizAttempt
	err = h.collection.FindOne(ctx, bson.M{
		"_id":        objectID,
		"student_id": studentID,
	}).Decode(&attempt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	if attempt.CompletedAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Attempt already completed"})
		return
	}

	now := time.Now()
	timeTaken := int(now.Sub(attempt.StartedAt).Seconds())

	update := bson.M{
		"$set": bson.M{
			"completed_at": now,
			"time_taken":   timeTaken,
		},
	}

	_, err = h.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete attempt"})
		return
	}

	// Get updated attempt
	err = h.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&attempt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attempt"})
		return
	}

	c.JSON(http.StatusOK, attempt)
}

// GetAttemptByID godoc
// @Summary      Get attempt by ID
// @Description  Get details of a specific quiz attempt
// @Tags         attempts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Attempt ID"
// @Success      200 {object} models.QuizAttempt
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /attempts/{id} [get]
func (h *AttemptHandler) GetAttemptByID(c *gin.Context) {
	attemptID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(attemptID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attempt ID"})
		return
	}

	userID, _ := c.Get("user_id")
	studentID := userID.(primitive.ObjectID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var attempt models.QuizAttempt
	err = h.collection.FindOne(ctx, bson.M{
		"_id":        objectID,
		"student_id": studentID,
	}).Decode(&attempt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	c.JSON(http.StatusOK, attempt)
}

// GetMyAttempts godoc
// @Summary      Get my attempts
// @Description  Get all quiz attempts by the authenticated student
// @Tags         attempts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.QuizAttempt
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /attempts [get]
func (h *AttemptHandler) GetMyAttempts(c *gin.Context) {
	userID, _ := c.Get("user_id")
	studentID := userID.(primitive.ObjectID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{primitive.E{Key: "started_at", Value: -1}})
	cursor, err := h.collection.Find(ctx, bson.M{"student_id": studentID}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attempts"})
		return
	}
	defer cursor.Close(ctx)

	var attempts []models.QuizAttempt
	if err := cursor.All(ctx, &attempts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode attempts"})
		return
	}

	c.JSON(http.StatusOK, attempts)
}
