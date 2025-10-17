// Package services provides business logic services
package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"quizmasterapi/config"
)

// CourseService handles external course API integration
type CourseService struct {
	BaseURL string
}

// NewCourseService creates a new course service
func NewCourseService() *CourseService {
	return &CourseService{
		BaseURL: config.AppConfig.ExternalCourseAPI,
	}
}

// CourseCompletionResponse represents the response from external course API
type CourseCompletionResponse struct {
	StudentID   string `json:"student_id"`
	CourseID    string `json:"course_id"`
	Completed   bool   `json:"completed"`
	CompletedAt string `json:"completed_at,omitempty"`
}

// CheckCourseCompletion checks if a student has completed a course
func (cs *CourseService) CheckCourseCompletion(studentID, courseID string) (bool, error) {
	// Build URL for external API
	url := fmt.Sprintf("%s/courses/%s/students/%s/completion", cs.BaseURL, courseID, studentID)

	// Make HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to call course API: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusNotFound {
		return false, nil // Student hasn't enrolled or completed
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("course API returned status %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	var completion CourseCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return false, fmt.Errorf("failed to parse response: %w", err)
	}

	return completion.Completed, nil
}
