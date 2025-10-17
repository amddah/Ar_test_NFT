# API Documentation - QuizMaster

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
All protected endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

---

## 1. Authentication Endpoints

### 1.1 Register User
**Endpoint:** `POST /auth/register`

**Description:** Register a new user (Professor or Student)

**Request Body:**
```json
{
  "email": "john.doe@university.com",
  "password": "securePassword123",
  "first_name": "John",
  "last_name": "Doe",
  "role": "student"
}
```

**Field Validations:**
- `email`: Valid email format, unique
- `password`: Minimum 6 characters
- `first_name`: Required
- `last_name`: Required
- `role`: Must be "professor" or "student"

**Success Response (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "64f8a9b2c3d4e5f6a7b8c9d0",
    "email": "john.doe@university.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "student",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Responses:**
- `400`: Invalid request body
- `409`: Email already registered

---

### 1.2 Login
**Endpoint:** `POST /auth/login`

**Description:** Authenticate user and receive JWT token

**Request Body:**
```json
{
  "email": "john.doe@university.com",
  "password": "securePassword123"
}
```

**Success Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "64f8a9b2c3d4e5f6a7b8c9d0",
    "email": "john.doe@university.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "student",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Responses:**
- `400`: Invalid request body
- `401`: Invalid email or password

---

## 2. User Management Endpoints

### 2.1 Get Profile
**Endpoint:** `GET /users/profile`

**Description:** Get current user's profile information

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "id": "64f8a9b2c3d4e5f6a7b8c9d0",
  "email": "john.doe@university.com",
  "first_name": "John",
  "last_name": "Doe",
  "role": "student",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `401`: Missing or invalid token
- `404`: User not found

---

## 3. Quiz Management Endpoints

### 3.1 Create Quiz
**Endpoint:** `POST /quizzes`

**Description:** Create a new quiz. Professors create approved quizzes, students create pending quizzes.

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "title": "Introduction to Algorithms",
  "description": "Test your understanding of basic algorithms",
  "category": "programming",
  "difficulty_level": "medium",
  "course_id": "CS201",
  "questions": [
    {
      "question_text": "What is the time complexity of binary search?",
      "type": "multiple_choice",
      "options": ["O(n)", "O(log n)", "O(n^2)", "O(1)"],
      "correct_answer": 1,
      "points": 15
    },
    {
      "question_text": "Is quicksort a stable sorting algorithm?",
      "type": "true_false",
      "correct_answer": false,
      "points": 10
    },
    {
      "question_text": "Which data structure uses LIFO principle?",
      "type": "multiple_choice",
      "options": ["Queue", "Stack", "Tree", "Graph"],
      "correct_answer": 1,
      "points": 10
    }
  ]
}
```

**Field Validations:**
- `title`: Required
- `category`: Must be one of: mathematics, science, history, literature, programming, languages
- `difficulty_level`: Must be one of: easy, medium, hard
- `course_id`: Required (external course reference)
- `questions`: Array with at least 1 question
- Question `type`: Must be "true_false" or "multiple_choice"
- Question `correct_answer`: Boolean for true_false, integer (option index) for multiple_choice
- Question `points`: Minimum 1

**Success Response (201):**
```json
{
  "id": "64f8a9b2c3d4e5f6a7b8c9d0",
  "title": "Introduction to Algorithms",
  "description": "Test your understanding of basic algorithms",
  "category": "programming",
  "difficulty_level": "medium",
  "course_id": "CS201",
  "creator_id": "64f8a9b2c3d4e5f6a7b8c9d1",
  "creator_role": "professor",
  "status": "approved",
  "questions": [...],
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400`: Invalid request body
- `401`: Unauthorized

---

### 3.2 Get All Quizzes
**Endpoint:** `GET /quizzes`

**Description:** Get list of quizzes with optional filters. Students only see approved quizzes.

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters:**
- `category` (optional): Filter by category
- `difficulty` (optional): Filter by difficulty level
- `status` (optional, professors only): Filter by status

**Example:**
```
GET /quizzes?category=programming&difficulty=medium
```

**Success Response (200):**
```json
[
  {
    "id": "64f8a9b2c3d4e5f6a7b8c9d0",
    "title": "Introduction to Algorithms",
    "description": "Test your understanding of basic algorithms",
    "category": "programming",
    "difficulty_level": "medium",
    "course_id": "CS201",
    "creator_id": "64f8a9b2c3d4e5f6a7b8c9d1",
    "status": "approved",
    "questions": [...],
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

---

### 3.3 Get Quiz by ID
**Endpoint:** `GET /quizzes/:id`

**Description:** Get detailed information about a specific quiz

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "id": "64f8a9b2c3d4e5f6a7b8c9d0",
  "title": "Introduction to Algorithms",
  "description": "Test your understanding of basic algorithms",
  "category": "programming",
  "difficulty_level": "medium",
  "course_id": "CS201",
  "creator_id": "64f8a9b2c3d4e5f6a7b8c9d1",
  "creator_role": "professor",
  "status": "approved",
  "questions": [
    {
      "id": "64f8a9b2c3d4e5f6a7b8c9d2",
      "question_text": "What is the time complexity of binary search?",
      "type": "multiple_choice",
      "options": ["O(n)", "O(log n)", "O(n^2)", "O(1)"],
      "correct_answer": 1,
      "time_limit": 15,
      "points": 15,
      "order": 1
    }
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400`: Invalid quiz ID format
- `403`: Quiz not available (students accessing pending quiz)
- `404`: Quiz not found

---

### 3.4 Approve/Reject Quiz
**Endpoint:** `PUT /quizzes/:id/approve`

**Description:** Approve or reject a student-created quiz (Professors only)

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "status": "approved"
}
```

**Field Validations:**
- `status`: Must be "approved" or "rejected"

**Success Response (200):**
```json
{
  "message": "Quiz status updated successfully"
}
```

**Error Responses:**
- `400`: Invalid request body or status value
- `401`: Unauthorized
- `403`: Insufficient permissions (not a professor)
- `404`: Quiz not found

---

### 3.5 Delete Quiz
**Endpoint:** `DELETE /quizzes/:id`

**Description:** Delete a quiz. Only creator or professors can delete.

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "message": "Quiz deleted successfully"
}
```

**Error Responses:**
- `400`: Invalid quiz ID
- `403`: No permission to delete
- `404`: Quiz not found

---

## 4. Quiz Attempt Endpoints (Students Only)

### 4.1 Start Quiz Attempt
**Endpoint:** `POST /attempts/start`

**Description:** Initiate a new quiz attempt. Validates course completion before allowing attempt.

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "quiz_id": "64f8a9b2c3d4e5f6a7b8c9d0"
}
```

**Success Response (201):**
```json
{
  "attempt": {
    "id": "64f8a9b2c3d4e5f6a7b8c9e0",
    "quiz_id": "64f8a9b2c3d4e5f6a7b8c9d0",
    "student_id": "64f8a9b2c3d4e5f6a7b8c9d1",
    "answers": [],
    "total_score": 0,
    "max_score": 35,
    "started_at": "2024-01-15T14:30:00Z"
  },
  "quiz": {
    "id": "64f8a9b2c3d4e5f6a7b8c9d0",
    "title": "Introduction to Algorithms",
    "questions": [
      {
        "id": "64f8a9b2c3d4e5f6a7b8c9d2",
        "question_text": "What is the time complexity of binary search?",
        "type": "multiple_choice",
        "options": ["O(n)", "O(log n)", "O(n^2)", "O(1)"],
        "time_limit": 15,
        "points": 15,
        "order": 1
      }
    ]
  }
}
```

**Note:** Questions are returned without `correct_answer` field during attempt.

**Error Responses:**
- `400`: Invalid quiz ID
- `403`: Course not completed or quiz not approved
- `404`: Quiz not found
- `409`: Already have an ongoing attempt

---

### 4.2 Submit Answer
**Endpoint:** `POST /attempts/answer`

**Description:** Submit an answer for a specific question in an ongoing attempt

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "attempt_id": "64f8a9b2c3d4e5f6a7b8c9e0",
  "question_id": "64f8a9b2c3d4e5f6a7b8c9d2",
  "answer": 1,
  "time_to_answer": 4
}
```

**Field Validations:**
- `attempt_id`: Valid ObjectID
- `question_id`: Valid ObjectID
- `answer`: Boolean (for true_false) or integer (for multiple_choice)
- `time_to_answer`: Integer, seconds taken to answer (max 15)

**Success Response (200):**
```json
{
  "is_correct": true,
  "points_earned": 15,
  "message": "Answer submitted successfully"
}
```

**Scoring Example:**
- Time to answer: 4 seconds (≤5s)
- Base points: 15
- Multiplier: 1.0 (100%)
- Points earned: 15

**Error Responses:**
- `400`: Invalid IDs, time limit exceeded, or answer already submitted
- `404`: Attempt or question not found

---

### 4.3 Complete Attempt
**Endpoint:** `PUT /attempts/:id/complete`

**Description:** Mark an attempt as complete and finalize scoring

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "id": "64f8a9b2c3d4e5f6a7b8c9e0",
  "quiz_id": "64f8a9b2c3d4e5f6a7b8c9d0",
  "student_id": "64f8a9b2c3d4e5f6a7b8c9d1",
  "answers": [
    {
      "question_id": "64f8a9b2c3d4e5f6a7b8c9d2",
      "student_answer": 1,
      "is_correct": true,
      "time_to_answer": 4,
      "points_earned": 15,
      "answered_at": "2024-01-15T14:31:00Z"
    },
    {
      "question_id": "64f8a9b2c3d4e5f6a7b8c9d3",
      "student_answer": false,
      "is_correct": true,
      "time_to_answer": 6,
      "points_earned": 8.8,
      "answered_at": "2024-01-15T14:32:00Z"
    }
  ],
  "total_score": 23.8,
  "max_score": 35,
  "started_at": "2024-01-15T14:30:00Z",
  "completed_at": "2024-01-15T14:33:00Z",
  "time_taken": 180
}
```

**Error Responses:**
- `400`: Invalid attempt ID or already completed
- `404`: Attempt not found

---

### 4.4 Get Attempt by ID
**Endpoint:** `GET /attempts/:id`

**Description:** Get details of a specific attempt

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "id": "64f8a9b2c3d4e5f6a7b8c9e0",
  "quiz_id": "64f8a9b2c3d4e5f6a7b8c9d0",
  "student_id": "64f8a9b2c3d4e5f6a7b8c9d1",
  "answers": [...],
  "total_score": 23.8,
  "max_score": 35,
  "started_at": "2024-01-15T14:30:00Z",
  "completed_at": "2024-01-15T14:33:00Z",
  "time_taken": 180
}
```

---

### 4.5 Get My Attempts
**Endpoint:** `GET /attempts`

**Description:** Get all attempts by the current user, sorted by most recent

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
[
  {
    "id": "64f8a9b2c3d4e5f6a7b8c9e0",
    "quiz_id": "64f8a9b2c3d4e5f6a7b8c9d0",
    "total_score": 23.8,
    "max_score": 35,
    "started_at": "2024-01-15T14:30:00Z",
    "completed_at": "2024-01-15T14:33:00Z"
  }
]
```

---

## 5. Leaderboard Endpoints

### 5.1 Get Quiz Leaderboard
**Endpoint:** `GET /leaderboards/quiz/:quiz_id`

**Description:** Get leaderboard for a specific quiz, sorted by score (desc) and time (asc)

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "quiz_id": "64f8a9b2c3d4e5f6a7b8c9d0",
  "total_count": 25,
  "leaderboard": [
    {
      "rank": 1,
      "student_id": "64f8a9b2c3d4e5f6a7b8c9e1",
      "student_name": "Alice Johnson",
      "score": 34.5,
      "max_score": 35,
      "percentage": 98.57,
      "time_taken": 120,
      "completed_at": "2024-01-15T15:00:00Z"
    },
    {
      "rank": 2,
      "student_id": "64f8a9b2c3d4e5f6a7b8c9e2",
      "student_name": "Bob Smith",
      "score": 34.5,
      "max_score": 35,
      "percentage": 98.57,
      "time_taken": 135,
      "completed_at": "2024-01-15T15:10:00Z"
    },
    {
      "rank": 3,
      "student_id": "64f8a9b2c3d4e5f6a7b8c9d1",
      "student_name": "John Doe",
      "score": 23.8,
      "max_score": 35,
      "percentage": 68,
      "time_taken": 180,
      "completed_at": "2024-01-15T14:33:00Z"
    }
  ]
}
```

---

### 5.2 Get My Rank
**Endpoint:** `GET /leaderboards/quiz/:quiz_id/my-rank`

**Description:** Get current user's rank for a specific quiz

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "quiz_id": "64f8a9b2c3d4e5f6a7b8c9d0",
  "rank": 3,
  "total_participants": 25,
  "score": 23.8,
  "max_score": 35,
  "percentage": 68,
  "time_taken": 180
}
```

**Error Responses:**
- `404`: No completed attempts found for this quiz

---

### 5.3 Get Global Leaderboard
**Endpoint:** `GET /leaderboards/global`

**Description:** Get top 50 performers across all quizzes based on average score

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "leaderboard": [
    {
      "rank": 1,
      "student_id": "64f8a9b2c3d4e5f6a7b8c9e1",
      "student_name": "Alice Johnson",
      "avg_score": 92.5,
      "total_attempts": 15,
      "total_score": 1387.5
    },
    {
      "rank": 2,
      "student_id": "64f8a9b2c3d4e5f6a7b8c9e2",
      "student_name": "Bob Smith",
      "avg_score": 89.3,
      "total_attempts": 12,
      "total_score": 1071.6
    }
  ]
}
```

---

## 6. Health Check

### 6.1 Health Check
**Endpoint:** `GET /health`

**Description:** Check if API is running

**Success Response (200):**
```json
{
  "status": "ok",
  "message": "QuizMaster API is running"
}
```

---

## Error Response Format

All error responses follow this format:

```json
{
  "error": "Detailed error message describing what went wrong"
}
```

**Common HTTP Status Codes:**
- `200`: Success
- `201`: Created
- `400`: Bad Request (invalid input)
- `401`: Unauthorized (missing/invalid token)
- `403`: Forbidden (insufficient permissions)
- `404`: Not Found
- `409`: Conflict (resource already exists)
- `500`: Internal Server Error

---

## Rate Limiting

Currently not implemented. Consider implementing rate limiting for production use.

## Pagination

Currently not implemented for list endpoints. All results are returned. Consider implementing pagination for production use with large datasets.

## CORS

Configure CORS settings based on your frontend application's domain.
