// Package services provides business logic services
package services

import (
	"math"
)

// ScoringService handles quiz scoring logic
type ScoringService struct{}

// NewScoringService creates a new scoring service
func NewScoringService() *ScoringService {
	return &ScoringService{}
}

// CalculateScore calculates points based on response time
// Rules:
// - Answered within 5 seconds → 100% of points
// - Answered between 5-10 seconds → 70% of points (linear decay)
// - Answered after 10 seconds → 50% of points
// - Maximum time is 15 seconds per question
func (ss *ScoringService) CalculateScore(basePoints int, timeToAnswer int, isCorrect bool) float64 {
	if !isCorrect {
		return 0
	}

	var multiplier float64

	if timeToAnswer <= 5 {
		// Fast answer - full points
		multiplier = 1.0
	} else if timeToAnswer <= 10 {
		// Medium speed - linear decay from 100% to 70%
		// Formula: 1.0 - ((timeToAnswer - 5) * 0.06)
		multiplier = 1.0 - ((float64(timeToAnswer-5) / 5.0) * 0.3)
	} else {
		// Slow answer - minimum 50% points
		multiplier = 0.5
	}

	score := float64(basePoints) * multiplier
	return math.Round(score*100) / 100 // Round to 2 decimal places
}

// CalculateTotalScore calculates the total score for all answers
func (ss *ScoringService) CalculateTotalScore(answers []struct {
	BasePoints   int
	TimeToAnswer int
	IsCorrect    bool
}) float64 {
	totalScore := 0.0
	for _, answer := range answers {
		score := ss.CalculateScore(answer.BasePoints, answer.TimeToAnswer, answer.IsCorrect)
		totalScore += score
	}
	return math.Round(totalScore*100) / 100
}
