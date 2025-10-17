// Package handlers provides HTTP request handlers
package handlers

import (
	"context"
	"net/http"
	"time"

	"quizmasterapi/config"
	"quizmasterapi/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// LeaderboardHandler handles leaderboard-related requests
type LeaderboardHandler struct {
	attemptCollection *mongo.Collection
	userCollection    *mongo.Collection
}

// NewLeaderboardHandler creates a new leaderboard handler
func NewLeaderboardHandler() *LeaderboardHandler {
	return &LeaderboardHandler{
		attemptCollection: config.GetCollection("attempts"),
		userCollection:    config.GetCollection("users"),
	}
}

// GetQuizLeaderboard godoc
// @Summary      Get quiz leaderboard
// @Description  Get the leaderboard for a specific quiz, showing rankings of all students
// @Tags         leaderboards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        quiz_id path string true "Quiz ID"
// @Success      200 {array} models.LeaderboardEntry
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /leaderboards/quiz/{quiz_id} [get]
func (h *LeaderboardHandler) GetQuizLeaderboard(c *gin.Context) {
	quizID := c.Param("quiz_id")
	objectID, err := primitive.ObjectIDFromHex(quizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find all completed attempts for this quiz, sorted by score (desc) and time (asc)
	filter := bson.M{
		"quiz_id":      objectID,
		"completed_at": bson.M{"$exists": true},
	}

	opts := options.Find().SetSort(bson.D{
		primitive.E{Key: "total_score", Value: -1},
		primitive.E{Key: "time_taken", Value: 1},
	})

	cursor, err := h.attemptCollection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}
	defer cursor.Close(ctx)

	var attempts []models.QuizAttempt
	if err := cursor.All(ctx, &attempts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode attempts"})
		return
	}

	// Build leaderboard with user information
	leaderboard := make([]models.LeaderboardEntry, 0, len(attempts))
	rank := 1

	for _, attempt := range attempts {
		// Get user info
		var user models.User
		err := h.userCollection.FindOne(ctx, bson.M{"_id": attempt.StudentID}).Decode(&user)
		if err != nil {
			continue // Skip if user not found
		}

		percentage := 0.0
		if attempt.MaxScore > 0 {
			percentage = (attempt.TotalScore / attempt.MaxScore) * 100
		}

		entry := models.LeaderboardEntry{
			Rank:        rank,
			StudentID:   attempt.StudentID,
			StudentName: user.FirstName + " " + user.LastName,
			Score:       attempt.TotalScore,
			MaxScore:    attempt.MaxScore,
			Percentage:  percentage,
			TimeTaken:   attempt.TimeTaken,
			CompletedAt: *attempt.CompletedAt,
		}

		leaderboard = append(leaderboard, entry)
		rank++
	}

	c.JSON(http.StatusOK, gin.H{
		"quiz_id":     quizID,
		"total_count": len(leaderboard),
		"leaderboard": leaderboard,
	})
}

// GetMyRank godoc
// @Summary      Get my rank
// @Description  Get the authenticated student's rank for a specific quiz
// @Tags         leaderboards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        quiz_id path string true "Quiz ID"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /leaderboards/quiz/{quiz_id}/my-rank [get]
func (h *LeaderboardHandler) GetMyRank(c *gin.Context) {
	quizID := c.Param("quiz_id")
	objectID, err := primitive.ObjectIDFromHex(quizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	userID, _ := c.Get("user_id")
	studentID := userID.(primitive.ObjectID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user's best attempt
	filter := bson.M{
		"quiz_id":      objectID,
		"student_id":   studentID,
		"completed_at": bson.M{"$exists": true},
	}

	opts := options.FindOne().SetSort(bson.D{
		primitive.E{Key: "total_score", Value: -1},
		primitive.E{Key: "time_taken", Value: 1},
	})

	var userAttempt models.QuizAttempt
	err = h.attemptCollection.FindOne(ctx, filter, opts).Decode(&userAttempt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No completed attempts found"})
		return
	}

	// Count how many attempts have better scores
	betterScoreFilter := bson.M{
		"quiz_id":      objectID,
		"completed_at": bson.M{"$exists": true},
		"$or": []bson.M{
			{"total_score": bson.M{"$gt": userAttempt.TotalScore}},
			{
				"total_score": userAttempt.TotalScore,
				"time_taken":  bson.M{"$lt": userAttempt.TimeTaken},
			},
		},
	}

	betterCount, err := h.attemptCollection.CountDocuments(ctx, betterScoreFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate rank"})
		return
	}

	rank := int(betterCount) + 1

	// Get total participants
	totalFilter := bson.M{
		"quiz_id":      objectID,
		"completed_at": bson.M{"$exists": true},
	}
	totalCount, err := h.attemptCollection.CountDocuments(ctx, totalFilter)
	if err != nil {
		totalCount = 0
	}

	percentage := 0.0
	if userAttempt.MaxScore > 0 {
		percentage = (userAttempt.TotalScore / userAttempt.MaxScore) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"quiz_id":            quizID,
		"rank":               rank,
		"total_participants": totalCount,
		"score":              userAttempt.TotalScore,
		"max_score":          userAttempt.MaxScore,
		"percentage":         percentage,
		"time_taken":         userAttempt.TimeTaken,
	})
}

// GetGlobalLeaderboard godoc
// @Summary      Get global leaderboard
// @Description  Get the top performing students across all quizzes (top 50)
// @Tags         leaderboards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /leaderboards/global [get]
func (h *LeaderboardHandler) GetGlobalLeaderboard(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregate pipeline to calculate average performance per student
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"completed_at": bson.M{"$exists": true}}}},
		{{Key: "$group", Value: bson.M{
			"_id": "$student_id",
			"avg_score": bson.M{"$avg": bson.M{
				"$multiply": bson.A{
					bson.M{"$divide": bson.A{"$total_score", "$max_score"}},
					100,
				},
			}},
			"total_attempts": bson.M{"$sum": 1},
			"total_score":    bson.M{"$sum": "$total_score"},
		}}},
		{{Key: "$sort", Value: bson.D{primitive.E{Key: "avg_score", Value: -1}}}},
		{{Key: "$limit", Value: 50}},
	}

	cursor, err := h.attemptCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate leaderboard"})
		return
	}
	defer cursor.Close(ctx)

	type GlobalEntry struct {
		StudentID     primitive.ObjectID `bson:"_id" json:"student_id"`
		AvgScore      float64            `bson:"avg_score" json:"avg_score"`
		TotalAttempts int                `bson:"total_attempts" json:"total_attempts"`
		TotalScore    float64            `bson:"total_score" json:"total_score"`
	}

	var results []GlobalEntry
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode results"})
		return
	}

	// Enrich with user information
	type GlobalLeaderboardEntry struct {
		Rank          int                `json:"rank"`
		StudentID     primitive.ObjectID `json:"student_id"`
		StudentName   string             `json:"student_name"`
		AvgScore      float64            `json:"avg_score"`
		TotalAttempts int                `json:"total_attempts"`
		TotalScore    float64            `json:"total_score"`
	}

	leaderboard := make([]GlobalLeaderboardEntry, 0, len(results))
	rank := 1

	for _, result := range results {
		var user models.User
		err := h.userCollection.FindOne(ctx, bson.M{"_id": result.StudentID}).Decode(&user)
		if err != nil {
			continue
		}

		entry := GlobalLeaderboardEntry{
			Rank:          rank,
			StudentID:     result.StudentID,
			StudentName:   user.FirstName + " " + user.LastName,
			AvgScore:      result.AvgScore,
			TotalAttempts: result.TotalAttempts,
			TotalScore:    result.TotalScore,
		}

		leaderboard = append(leaderboard, entry)
		rank++
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
	})
}
