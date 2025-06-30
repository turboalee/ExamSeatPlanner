package seating

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SeatingHandler handles HTTP requests for seating operations.
type SeatingHandler struct {
	service *SeatingService
}

// NewSeatingHandler creates a new SeatingHandler.
func NewSeatingHandler(service *SeatingService) *SeatingHandler {
	return &SeatingHandler{service: service}
}

// GenerateSeatingPlanRequest represents the request to generate a seating plan.
type GenerateSeatingPlanRequest struct {
	ExamID           string   `json:"exam_id"`           // Exam ID
	RoomID           string   `json:"room_id"`           // Room ID
	InvigilatorEmail string   `json:"invigilator_email"` // Invigilator email
	Algorithm        string   `json:"algorithm"`         // Algorithm to use (matrix, parallel, random)
	StudentIDs       []string `json:"student_ids"`       // List of student IDs
}

// CreateExamRequest represents the request to create an exam.
type CreateExamRequest struct {
	Title         string    `json:"title"`          // Exam title
	Date          time.Time `json:"date"`           // Exam date
	Duration      int       `json:"duration"`       // Duration in minutes
	Faculty       string    `json:"faculty"`        // Faculty
	Department    string    `json:"department"`     // Department
	Course        string    `json:"course"`         // Course code
	Batch         string    `json:"batch"`          // Batch
	TotalStudents int       `json:"total_students"` // Total students
}

// CreateRoomRequest represents the request to create a room.
type CreateRoomRequest struct {
	Name     string `json:"name"`     // Room name
	Capacity int    `json:"capacity"` // Total capacity
	Rows     int    `json:"rows"`     // Number of rows
	Columns  int    `json:"columns"`  // Number of columns
	Building string `json:"building"` // Building name
}

// CreateStudentRequest represents the request to create a student.
type CreateStudentRequest struct {
	StudentID  string `json:"student_id"` // Student ID
	Name       string `json:"name"`       // Student name
	Email      string `json:"email"`      // Student email
	Department string `json:"department"` // Department
	Batch      string `json:"batch"`      // Batch
	Course     string `json:"course"`     // Course
	Faculty    string `json:"faculty"`    // Faculty
}

// CreateInvigilatorRequest represents the request to create an invigilator.
type CreateInvigilatorRequest struct {
	Email   string `json:"email"`   // Invigilator email
	Name    string `json:"name"`    // Invigilator name
	Faculty string `json:"faculty"` // Faculty
}

// GenerateSeatingPlan allows admins to generate a new seating plan.
func (h *SeatingHandler) GenerateSeatingPlan(c echo.Context) error {
	var req GenerateSeatingPlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate algorithm
	if req.Algorithm != "matrix" && req.Algorithm != "parallel" && req.Algorithm != "random" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid algorithm. Must be matrix, parallel, or random"})
	}

	// Convert string IDs to ObjectIDs
	examID, err := primitive.ObjectIDFromHex(req.ExamID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid exam ID"})
	}

	roomID, err := primitive.ObjectIDFromHex(req.RoomID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid room ID"})
	}

	studentIDs := make([]primitive.ObjectID, len(req.StudentIDs))
	for i, studentIDStr := range req.StudentIDs {
		studentID, err := primitive.ObjectIDFromHex(studentIDStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid student ID"})
		}
		studentIDs[i] = studentID
	}

	plan, err := h.service.GenerateSeatingPlan(context.Background(), examID, roomID, req.InvigilatorEmail, req.Algorithm, studentIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, plan)
}

// GetSeatingPlan retrieves a seating plan by ID.
func (h *SeatingHandler) GetSeatingPlan(c echo.Context) error {
	planID := c.Param("id")
	if planID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Plan ID is required"})
	}

	id, err := primitive.ObjectIDFromHex(planID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid plan ID"})
	}

	plan, err := h.service.GetSeatingPlan(context.Background(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if plan == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Seating plan not found"})
	}

	return c.JSON(http.StatusOK, plan)
}

// CreateExam allows admins to create a new exam.
func (h *SeatingHandler) CreateExam(c echo.Context) error {
	var req CreateExamRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	exam := &Exam{
		ID:            primitive.NewObjectID(),
		Title:         req.Title,
		Date:          req.Date,
		Duration:      req.Duration,
		Faculty:       req.Faculty,
		Department:    req.Department,
		Course:        req.Course,
		Batch:         req.Batch,
		TotalStudents: req.TotalStudents,
	}

	err := h.service.repo.CreateExam(context.Background(), exam)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create exam"})
	}

	return c.JSON(http.StatusCreated, exam)
}

// CreateRoom allows admins to create a new room.
func (h *SeatingHandler) CreateRoom(c echo.Context) error {
	var req CreateRoomRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	room := &Room{
		ID:       primitive.NewObjectID(),
		Name:     req.Name,
		Capacity: req.Capacity,
		Rows:     req.Rows,
		Columns:  req.Columns,
		Building: req.Building,
	}

	err := h.service.repo.CreateRoom(context.Background(), room)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create room"})
	}

	return c.JSON(http.StatusCreated, room)
}

// CreateStudent allows staff to create a new student.
func (h *SeatingHandler) CreateStudent(c echo.Context) error {
	var req CreateStudentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	student := &Student{
		ID:         primitive.NewObjectID(),
		StudentID:  req.StudentID,
		Name:       req.Name,
		Email:      req.Email,
		Department: req.Department,
		Batch:      req.Batch,
		Course:     req.Course,
		Faculty:    req.Faculty,
	}

	err := h.service.repo.CreateStudent(context.Background(), student)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create student"})
	}

	return c.JSON(http.StatusCreated, student)
}

// CreateInvigilator allows admins to create a new invigilator.
func (h *SeatingHandler) CreateInvigilator(c echo.Context) error {
	var req CreateInvigilatorRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	invigilator := &Invigilator{
		ID:      primitive.NewObjectID(),
		Email:   req.Email,
		Name:    req.Name,
		Faculty: req.Faculty,
	}

	err := h.service.repo.CreateInvigilator(context.Background(), invigilator)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create invigilator"})
	}

	return c.JSON(http.StatusCreated, invigilator)
}

// Why: This handler provides HTTP interfaces for all seating-related operations, with proper validation and error handling for each endpoint.
