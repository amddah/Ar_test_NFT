# QuizMaster API

A comprehensive backend API for a mobile quiz application built with Go (Gin framework) and MongoDB. This API supports role-based access control, gamified scoring, external course integration, and competitive leaderboards.

## 🚀 Features

- **User Management**: Registration and authentication with role-based access (Professor/Student)
- **Quiz Management**: Create, approve, and manage quizzes across multiple categories and difficulty levels
- **Dynamic Question Types**: Support for True/False and Multiple Choice questions
- **Time-Based Scoring**: Advanced scoring algorithm based on response time
- **Course Integration**: Integration with external course completion API
- **Quiz Attempts**: Track student attempts with detailed answer recording
- **Leaderboards**: Quiz-specific and global leaderboards for competitive gameplay
- **Approval Workflow**: Student-created quizzes require professor approval

## 📋 Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Data Models](#data-models)
- [API Endpoints](#api-endpoints)
- [Scoring System](#scoring-system)
- [External Course Integration](#external-course-integration)
- [Error Handling](#error-handling)

## 🔧 Requirements

- Go 1.23 or higher
- MongoDB 4.4 or higher
- External Course API (for course completion verification)

## 📦 Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd quizmasterapi
```

2. **Install dependencies**
```bash
go mod download
```

3. **Configure environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Run the application**
```bash
go run main.go
```

The server will start on `http://localhost:8080` (or the port specified in your .env file)

## ⚙️ Configuration

Create a `.env` file in the root directory with the following variables:

```env
MONGO_URI=mongodb://localhost:27017
DATABASE_NAME=quizmaster
JWT_SECRET=your-secret-key-change-in-production
SERVER_PORT=8080
EXTERNAL_COURSE_API=http://localhost:9000/api/v1
```

## 📊 Data Models

### User
```json
{
  "id": "ObjectId",
  "email": "string",
  "first_name": "string",
  "last_name": "string",
  "role": "professor|student",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Quiz
```json
{
  "id": "ObjectId",
  "title": "string",
  "description": "string",
  "category": "mathematics|science|history|literature|programming|languages",
  "difficulty_level": "easy|medium|hard",
  "course_id": "string",
  "creator_id": "ObjectId",
  "creator_role": "professor|student",
  "status": "pending|approved|rejected",
  "questions": [
    {
      "id": "ObjectId",
      "question_text": "string",
      "type": "true_false|multiple_choice",
      "options": ["string"],
      "correct_answer": "bool|int",
      "time_limit": 15,
      "points": "int",
      "order": "int"
    }
  ],
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "approved_by": "ObjectId",
  "approved_at": "timestamp"
}
```

### Quiz Attempt
```json
{
  "id": "ObjectId",
  "quiz_id": "ObjectId",
  "student_id": "ObjectId",
  "answers": [
    {
      "question_id": "ObjectId",
      "student_answer": "bool|int",
      "is_correct": "bool",
      "time_to_answer": "int",
      "points_earned": "float",
      "answered_at": "timestamp"
    }
  ],
  "total_score": "float",
  "max_score": "float",
  "started_at": "timestamp",
  "completed_at": "timestamp",
  "time_taken": "int"
}
```

### Leaderboard Entry
```json
{
  "rank": "int",
  "student_id": "ObjectId",
  "student_name": "string",
  "score": "float",
  "max_score": "float",
  "percentage": "float",
  "time_taken": "int",
  "completed_at": "timestamp"
}
```

## 🛣️ API Endpoints

### Authentication

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe",
  "role": "student"
}

Response 201:
{
  "token": "jwt-token",
  "user": { ... }
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}

Response 200:
{
  "token": "jwt-token",
  "user": { ... }
}
```

### User Management

#### Get Profile
```http
GET /api/v1/users/profile
Authorization: Bearer <token>

Response 200:
{
  "id": "...",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "role": "student",
  ...
}
```

### Quiz Management

#### Create Quiz
```http
POST /api/v1/quizzes
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Basic Mathematics",
  "description": "Test your math skills",
  "category": "mathematics",
  "difficulty_level": "easy",
  "course_id": "MATH101",
  "questions": [
    {
      "question_text": "What is 2 + 2?",
      "type": "multiple_choice",
      "options": ["3", "4", "5", "6"],
      "correct_answer": 1,
      "points": 10
    },
    {
      "question_text": "Is 10 > 5?",
      "type": "true_false",
      "correct_answer": true,
      "points": 5
    }
  ]
}

Response 201:
{
  "id": "...",
  "title": "Basic Mathematics",
  "status": "approved",  // "pending" for students
  ...
}
```

#### Get All Quizzes
```http
GET /api/v1/quizzes?category=mathematics&difficulty=easy&status=approved
Authorization: Bearer <token>

Response 200:
[
  {
    "id": "...",
    "title": "Basic Mathematics",
    ...
  }
]
```

#### Get Quiz by ID
```http
GET /api/v1/quizzes/:id
Authorization: Bearer <token>

Response 200:
{
  "id": "...",
  "title": "Basic Mathematics",
  "questions": [...],
  ...
}
```

#### Approve/Reject Quiz (Professor Only)
```http
PUT /api/v1/quizzes/:id/approve
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "approved"
}

Response 200:
{
  "message": "Quiz status updated successfully"
}
```

#### Delete Quiz
```http
DELETE /api/v1/quizzes/:id
Authorization: Bearer <token>

Response 200:
{
  "message": "Quiz deleted successfully"
}
```

### Quiz Attempts (Student Only)

#### Start Quiz Attempt
```http
POST /api/v1/attempts/start
Authorization: Bearer <token>
Content-Type: application/json

{
  "quiz_id": "64f8a9b2c3d4e5f6a7b8c9d0"
}

Response 201:
{
  "attempt": {
    "id": "...",
    "quiz_id": "...",
    "student_id": "...",
    "max_score": 100,
    "started_at": "..."
  },
  "quiz": {
    "id": "...",
    "title": "...",
    "questions": [...]  // Without correct answers
  }
}
```

#### Submit Answer
```http
POST /api/v1/attempts/answer
Authorization: Bearer <token>
Content-Type: application/json

{
  "attempt_id": "64f8a9b2c3d4e5f6a7b8c9d0",
  "question_id": "64f8a9b2c3d4e5f6a7b8c9d1",
  "answer": 1,
  "time_to_answer": 3
}

Response 200:
{
  "is_correct": true,
  "points_earned": 10,
  "message": "Answer submitted successfully"
}
```

#### Complete Attempt
```http
PUT /api/v1/attempts/:id/complete
Authorization: Bearer <token>

Response 200:
{
  "id": "...",
  "quiz_id": "...",
  "student_id": "...",
  "total_score": 85,
  "max_score": 100,
  "completed_at": "...",
  "time_taken": 120,
  "answers": [...]
}
```

#### Get Attempt by ID
```http
GET /api/v1/attempts/:id
Authorization: Bearer <token>

Response 200:
{
  "id": "...",
  "quiz_id": "...",
  "answers": [...],
  "total_score": 85,
  ...
}
```

#### Get My Attempts
```http
GET /api/v1/attempts
Authorization: Bearer <token>

Response 200:
[
  {
    "id": "...",
    "quiz_id": "...",
    "total_score": 85,
    ...
  }
]
```

### Leaderboards

#### Get Quiz Leaderboard
```http
GET /api/v1/leaderboards/quiz/:quiz_id
Authorization: Bearer <token>

Response 200:
{
  "quiz_id": "...",
  "total_count": 50,
  "leaderboard": [
    {
      "rank": 1,
      "student_id": "...",
      "student_name": "John Doe",
      "score": 95,
      "max_score": 100,
      "percentage": 95,
      "time_taken": 90,
      "completed_at": "..."
    }
  ]
}
```

#### Get My Rank
```http
GET /api/v1/leaderboards/quiz/:quiz_id/my-rank
Authorization: Bearer <token>

Response 200:
{
  "quiz_id": "...",
  "rank": 5,
  "total_participants": 50,
  "score": 85,
  "max_score": 100,
  "percentage": 85,
  "time_taken": 120
}
```

#### Get Global Leaderboard
```http
GET /api/v1/leaderboards/global
Authorization: Bearer <token>

Response 200:
{
  "leaderboard": [
    {
      "rank": 1,
      "student_id": "...",
      "student_name": "Jane Smith",
      "avg_score": 92.5,
      "total_attempts": 15,
      "total_score": 1387.5
    }
  ]
}
```

## 🎯 Scoring System

The scoring system is time-based to encourage quick thinking:

### Scoring Rules

- **Fast Answer (≤5 seconds)**: 100% of base points
- **Medium Answer (5-10 seconds)**: 70-100% of base points (linear decay)
- **Slow Answer (>10 seconds)**: 50% of base points
- **Maximum Time**: 15 seconds per question
- **Wrong Answer**: 0 points

### Scoring Formula

```
if time_to_answer <= 5:
    points = base_points * 1.0

else if time_to_answer <= 10:
    multiplier = 1.0 - ((time_to_answer - 5) / 5.0) * 0.3
    points = base_points * multiplier

else:
    points = base_points * 0.5
```

### Example Scoring

For a question worth 10 base points:
- Answered in 3 seconds: **10 points** (100%)
- Answered in 7 seconds: **8.8 points** (88%)
- Answered in 12 seconds: **5 points** (50%)

## 🔗 External Course Integration

The API integrates with an external course service to verify course completion before allowing quiz attempts.

### Expected External API Format

```http
GET {EXTERNAL_COURSE_API}/courses/{course_id}/students/{student_id}/completion

Response 200:
{
  "student_id": "64f8a9b2c3d4e5f6a7b8c9d0",
  "course_id": "MATH101",
  "completed": true,
  "completed_at": "2024-01-15T10:30:00Z"
}

Response 404:
Student not enrolled or hasn't completed the course
```

### Integration Flow

1. Student attempts to start a quiz
2. API retrieves quiz details and associated `course_id`
3. API calls external course service to verify completion
4. If completed, attempt is allowed; otherwise, returns 403 Forbidden

## ⚠️ Error Handling

The API uses standard HTTP status codes:

- **200 OK**: Successful request
- **201 Created**: Resource created successfully
- **400 Bad Request**: Invalid request payload
- **401 Unauthorized**: Missing or invalid authentication token
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource not found
- **409 Conflict**: Resource already exists
- **500 Internal Server Error**: Server error

### Error Response Format

```json
{
  "error": "Detailed error message"
}
```

## 🔐 Authentication

All protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer <jwt-token>
```

Tokens are issued upon registration or login and are valid for 24 hours.

## 🏗️ Project Structure

```
quizmasterapi/
├── config/              # Configuration and database setup
│   └── config.go
├── models/              # Data models
│   └── models.go
├── middleware/          # HTTP middleware (auth, CORS, etc.)
│   └── auth.go
├── handlers/            # HTTP request handlers
│   ├── user_handler.go
│   ├── quiz_handler.go
│   ├── attempt_handler.go
│   └── leaderboard_handler.go
├── services/            # Business logic services
│   ├── course_service.go
│   └── scoring_service.go
├── main.go              # Application entry point
├── go.mod               # Go module dependencies
├── .env.example         # Example environment configuration
└── README.md            # This file
```

## 🧪 Testing

### Manual Testing with cURL

#### 1. Register a student
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "student@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Student",
    "role": "student"
  }'
```

#### 2. Register a professor
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "prof@example.com",
    "password": "password123",
    "first_name": "Jane",
    "last_name": "Professor",
    "role": "professor"
  }'
```

#### 3. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "student@example.com",
    "password": "password123"
  }'
```

#### 4. Create a quiz (as professor)
```bash
curl -X POST http://localhost:8080/api/v1/quizzes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Basic Math Quiz",
    "description": "Test your basic math skills",
    "category": "mathematics",
    "difficulty_level": "easy",
    "course_id": "MATH101",
    "questions": [
      {
        "question_text": "What is 5 + 3?",
        "type": "multiple_choice",
        "options": ["6", "7", "8", "9"],
        "correct_answer": 2,
        "points": 10
      },
      {
        "question_text": "Is 10 an even number?",
        "type": "true_false",
        "correct_answer": true,
        "points": 5
      }
    ]
  }'
```

## 📝 License

This project is licensed under the MIT License.

## 👥 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📞 Support

For support, email support@quizmaster.com or open an issue in the repository.

---

**Built with ❤️ using Go, Gin, and MongoDB**
