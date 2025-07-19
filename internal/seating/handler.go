package seating

import (
	"context"
	"log"
	"net/http"
	"reflect"
	"time"

	"ExamSeatPlanner/internal/auth"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
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
	Title     string    `json:"title"`     // Exam title
	Date      time.Time `json:"date"`      // Exam date
	Duration  int       `json:"duration"`  // Duration in minutes
	Faculty   string    `json:"faculty"`   // Faculty
	Algorithm string    `json:"algorithm"` // Preferred seating algorithm
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

// UploadStudentListRequest represents the request to upload a student list.
type UploadStudentListRequest struct {
	Name       string    `json:"name"`        // Name of the list (e.g., "BSIT 2022")
	Department string    `json:"department"`  // Department
	Course     string    `json:"course"`      // Course
	Batch      string    `json:"batch"`       // Batch
	Faculty    string    `json:"faculty"`     // Faculty
	Students   []Student `json:"students"`    // List of students
	UploadedBy string    `json:"uploaded_by"` // ID of the staff who uploaded
}

// AddRoomToExamRequest represents the request to add a room to an exam.
type AddRoomToExamRequest struct {
	ExamID         string   `json:"exam_id"`          // Exam ID
	RoomID         string   `json:"room_id"`          // Room ID
	StudentListIDs []string `json:"student_list_ids"` // Student list IDs to assign to this room
}

// AddInvigilatorToRoomRequest represents the request to add an invigilator to a room.
type AddInvigilatorToRoomRequest struct {
	ExamRoomID    string `json:"exam_room_id"`   // Exam room ID
	InvigilatorID string `json:"invigilator_id"` // Invigilator ID
}

// GenerateSeatingPlan allows admins to generate a new seating plan.
func (h *SeatingHandler) GenerateSeatingPlan(c echo.Context) error {
	var req GenerateSeatingPlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate algorithm
	if req.Algorithm != "parallel" && req.Algorithm != "simple" && req.Algorithm != "separated" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid algorithm. Must be 'parallel', 'simple', or 'separated'"})
	}

	// Convert string IDs to ObjectIDs
	examID, err := primitive.ObjectIDFromHex(req.ExamID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid exam ID"})
	}

	plans, err := h.service.GenerateSeatingPlan(context.Background(), examID, primitive.NilObjectID, req.InvigilatorEmail, req.Algorithm, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, plans)
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
		// Log the error for debugging
		log.Printf("[CreateExam] Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request: " + err.Error()})
	}

	exam := &Exam{
		ID:        primitive.NewObjectID(),
		Title:     req.Title,
		Date:      req.Date,
		Duration:  req.Duration,
		Faculty:   req.Faculty,
		Algorithm: req.Algorithm,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := h.service.repo.CreateExam(context.Background(), exam)
	if err != nil {
		// Log the error for debugging
		log.Printf("[CreateExam] Failed to create exam: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create exam: " + err.Error()})
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
		StudentID: req.StudentID,
		Name:      req.Name,
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

// UploadStudentList handles uploading a new student list
func (h *SeatingHandler) UploadStudentList(c echo.Context) error {
	var req struct {
		Department string    `json:"department"`
		Batch      string    `json:"batch"`
		Faculty    string    `json:"faculty"`
		Students   []Student `json:"students"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if req.Department == "" || req.Batch == "" || req.Faculty == "" || len(req.Students) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields"})
	}
	// Robustly extract email from JWT claims (map or struct)
	user := c.Get("user")
	var uploadedBy string
	switch u := user.(type) {
	case map[string]interface{}:
		if email, ok := u["email"].(string); ok && email != "" {
			uploadedBy = email
		}
	default:
		// Try reflection for struct with Email field
		v := reflect.ValueOf(user)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			emailField := v.FieldByName("Email")
			if emailField.IsValid() && emailField.Kind() == reflect.String {
				email := emailField.String()
				if email != "" {
					uploadedBy = email
				}
			}
		}
	}
	if uploadedBy == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Could not determine uploader from authentication context"})
	}
	// Only keep student_id and name for each student
	students := make([]Student, 0, len(req.Students))
	for _, s := range req.Students {
		students = append(students, Student{
			StudentID: s.StudentID,
			Name:      s.Name,
		})
	}
	// Auto-generate list name as Department/Batch
	listName := req.Department + "/" + req.Batch
	studentList := StudentList{
		Department: req.Department,
		Batch:      req.Batch,
		Faculty:    req.Faculty,
		Name:       listName,
		Students:   students,
		UploadedBy: uploadedBy,
	}
	if err := h.service.repo.CreateStudentList(c.Request().Context(), &studentList); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save student list"})
	}
	// Insert each student into the students collection if not already present
	for _, s := range students {
		existing, _ := h.service.repo.FindStudentByID(c.Request().Context(), s.StudentID)
		if existing == nil {
			h.service.repo.CreateStudent(c.Request().Context(), &s)
		}
	}
	return c.JSON(http.StatusOK, studentList)
}

// AddRoomToExam allows admins to add a room to an exam.
func (h *SeatingHandler) AddRoomToExam(c echo.Context) error {
	log.Printf("[AddRoomToExam] Handler called")
	var req AddRoomToExamRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[AddRoomToExam] Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request: " + err.Error()})
	}

	examID, err := primitive.ObjectIDFromHex(req.ExamID)
	if err != nil {
		log.Printf("[AddRoomToExam] Invalid exam ID: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid exam ID: " + err.Error()})
	}

	roomID, err := primitive.ObjectIDFromHex(req.RoomID)
	if err != nil {
		log.Printf("[AddRoomToExam] Invalid room ID: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid room ID: " + err.Error()})
	}

	// Parse all student list IDs
	var studentListObjIDs []primitive.ObjectID
	for _, idStr := range req.StudentListIDs {
		objID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			log.Printf("[AddRoomToExam] Invalid student list ID: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid student list ID: " + err.Error()})
		}
		studentListObjIDs = append(studentListObjIDs, objID)
	}

	examRoom := &ExamRoom{
		ID:             primitive.NewObjectID(),
		ExamID:         examID,
		RoomID:         roomID,
		StudentListIDs: studentListObjIDs,
		Invigilators:   []primitive.ObjectID{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = h.service.repo.CreateExamRoom(context.Background(), examRoom)
	if err != nil {
		log.Printf("[AddRoomToExam] Failed to add room to exam: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add room to exam: " + err.Error()})
	}

	return c.JSON(http.StatusCreated, examRoom)
}

// AddInvigilatorToRoom allows admins to add an invigilator to a room.
func (h *SeatingHandler) AddInvigilatorToRoom(c echo.Context) error {
	var req AddInvigilatorToRoomRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[AddInvigilatorToRoom] Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	log.Printf("[AddInvigilatorToRoom] Parsed request: %+v", req)
	if req.ExamRoomID == "" || req.InvigilatorID == "" {
		log.Printf("[AddInvigilatorToRoom] Missing exam_room_id or invigilator_id")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing exam_room_id or invigilator_id"})
	}

	examRoomID, err := primitive.ObjectIDFromHex(req.ExamRoomID)
	if err != nil {
		log.Printf("[AddInvigilatorToRoom] Invalid exam room ID: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid exam room ID"})
	}

	invigilatorID, err := primitive.ObjectIDFromHex(req.InvigilatorID)
	if err != nil {
		log.Printf("[AddInvigilatorToRoom] Invalid invigilator ID: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invigilator ID"})
	}

	// Check if invigilator is already assigned to another room in the same exam
	examRoom, err := h.service.repo.FindExamRoomByID(context.Background(), examRoomID)
	if err != nil {
		log.Printf("[AddInvigilatorToRoom] Failed to find exam room: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find exam room"})
	}
	if examRoom == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Exam room not found"})
	}

	// Get all exam rooms for this exam
	examRooms, err := h.service.repo.GetExamRooms(context.Background(), examRoom.ExamID)
	if err != nil {
		log.Printf("[AddInvigilatorToRoom] Failed to get exam rooms: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get exam rooms"})
	}

	// Check if invigilator is already assigned to any other room in this exam
	for _, er := range examRooms {
		if er.ID != examRoomID { // Skip current room
			for _, existingInvID := range er.Invigilators {
				if existingInvID == invigilatorID {
					return c.JSON(http.StatusConflict, map[string]string{"error": "Invigilator is already assigned to another room in this exam"})
				}
			}
		}
	}

	log.Printf("[AddInvigilatorToRoom] Assigning invigilator %s to exam room %s", invigilatorID.Hex(), examRoomID.Hex())
	err = h.service.repo.AddInvigilatorToRoom(context.Background(), examRoomID, invigilatorID)
	if err != nil {
		log.Printf("[AddInvigilatorToRoom] Failed to add invigilator: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add invigilator to room"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Invigilator added to room successfully"})
}

// DeleteExam allows admins to delete an exam by ID.
func (h *SeatingHandler) DeleteExam(c echo.Context) error {
	examID := c.Param("id")
	if examID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Exam ID is required"})
	}

	id, err := primitive.ObjectIDFromHex(examID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid exam ID"})
	}

	err = h.service.repo.DeleteExam(context.Background(), id)
	if err != nil {
		log.Printf("[DeleteExam] Failed to delete exam: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete exam: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Exam deleted successfully"})
}

// UpdateExam allows admins to update an exam by ID.
func (h *SeatingHandler) UpdateExam(c echo.Context) error {
	examID := c.Param("id")
	if examID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Exam ID is required"})
	}

	id, err := primitive.ObjectIDFromHex(examID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid exam ID"})
	}

	var req CreateExamRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[UpdateExam] Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request: " + err.Error()})
	}

	exam := &Exam{
		ID:        id,
		Title:     req.Title,
		Date:      req.Date,
		Duration:  req.Duration,
		Faculty:   req.Faculty,
		Algorithm: req.Algorithm,
		UpdatedAt: time.Now(),
	}

	err = h.service.repo.UpdateExam(context.Background(), exam)
	if err != nil {
		log.Printf("[UpdateExam] Failed to update exam: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update exam: " + err.Error()})
	}

	return c.JSON(http.StatusOK, exam)
}

// UpdateRoom allows admins to update a room by ID.
func (h *SeatingHandler) UpdateRoom(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing room ID"})
	}
	roomID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid room ID"})
	}

	var room Room
	if err := c.Bind(&room); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	err = h.service.UpdateRoom(c.Request().Context(), roomID, &room)
	if err != nil {
		if err.Error() == "room not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Room not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update room"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Room updated successfully"})
}

// GetAllExams retrieves all exams.
func (h *SeatingHandler) GetAllExams(c echo.Context) error {
	exams, err := h.service.repo.GetAllExams(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch exams"})
	}
	return c.JSON(http.StatusOK, exams)
}

// GetAllStudents retrieves all students.
func (h *SeatingHandler) GetAllStudents(c echo.Context) error {
	students, err := h.service.repo.GetAllStudents(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch students"})
	}
	// Debug log: print all students being returned
	log.Printf("[GetAllStudents] Returning %d students. Sample: %+v", len(students), func() interface{} {
		if len(students) > 0 {
			return students[0]
		} else {
			return nil
		}
	}())
	return c.JSON(http.StatusOK, students)
}

// GetAllSeatingPlans retrieves all seating plans.
func (h *SeatingHandler) GetAllSeatingPlans(c echo.Context) error {
	plans, err := h.service.repo.GetAllSeatingPlans(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch seating plans"})
	}
	return c.JSON(http.StatusOK, plans)
}

// GetAllRooms retrieves all rooms.
func (h *SeatingHandler) GetAllRooms(c echo.Context) error {
	rooms, err := h.service.repo.GetAllRooms(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch rooms"})
	}
	return c.JSON(http.StatusOK, rooms)
}

// GetAllStudentLists retrieves all student lists.
func (h *SeatingHandler) GetAllStudentLists(c echo.Context) error {
	studentLists, err := h.service.repo.GetAllStudentLists(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch student lists"})
	}
	return c.JSON(http.StatusOK, studentLists)
}

// Add after GetAllStudentLists
func (h *SeatingHandler) DeleteStudentList(c echo.Context) error {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}
	if err := h.service.repo.DeleteStudentList(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete student list"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *SeatingHandler) UpdateStudentList(c echo.Context) error {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}
	var update bson.M
	if err := c.Bind(&update); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	if err := h.service.repo.UpdateStudentList(c.Request().Context(), id, update); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update student list"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Add a student to a student list
func (h *SeatingHandler) AddStudentToList(c echo.Context) error {
	listIDStr := c.Param("id")
	listID, err := primitive.ObjectIDFromHex(listIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid list ID"})
	}
	var student Student
	if err := c.Bind(&student); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	if student.StudentID == "" || student.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Student ID and Name are required"})
	}
	if err := h.service.repo.AddStudentToList(c.Request().Context(), listID, student); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add student"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Update a student in a student list
func (h *SeatingHandler) UpdateStudentInList(c echo.Context) error {
	listIDStr := c.Param("id")
	studentID := c.Param("studentId")
	listID, err := primitive.ObjectIDFromHex(listIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid list ID"})
	}
	var student Student
	if err := c.Bind(&student); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	if student.StudentID == "" || student.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Student ID and Name are required"})
	}
	if err := h.service.repo.UpdateStudentInList(c.Request().Context(), listID, studentID, student); err != nil {
		if err.Error() == "student_id already exists in this list" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update student"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Remove a student from a student list
func (h *SeatingHandler) RemoveStudentFromList(c echo.Context) error {
	listIDStr := c.Param("id")
	studentID := c.Param("studentId")
	listID, err := primitive.ObjectIDFromHex(listIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid list ID"})
	}
	if err := h.service.repo.RemoveStudentFromList(c.Request().Context(), listID, studentID); err != nil {
		if err.Error() == "student not found in list" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to remove student"})
	}
	return c.NoContent(http.StatusNoContent)
}

// GetAllInvigilators retrieves all invigilators (now users with role admin or staff)
func (h *SeatingHandler) GetAllInvigilators(c echo.Context) error {
	invigilators, err := h.service.repo.GetAllInvigilators(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch invigilators"})
	}
	return c.JSON(http.StatusOK, invigilators)
}

// GetExamRooms retrieves all rooms for a specific exam.
func (h *SeatingHandler) GetExamRooms(c echo.Context) error {
	examID := c.Param("examId")
	if examID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Exam ID is required"})
	}

	id, err := primitive.ObjectIDFromHex(examID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid exam ID"})
	}

	examRooms, err := h.service.repo.GetExamRooms(context.Background(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch exam rooms"})
	}

	// For each examRoom, fetch room, student list, and invigilator details, with debug logs
	var result []map[string]interface{}
	for _, er := range examRooms {
		log.Printf("[GetExamRooms] ExamRoom: %v", er)
		room, _ := h.service.repo.FindRoomByID(context.Background(), er.RoomID)
		if room == nil {
			log.Printf("[GetExamRooms] Room not found for ID: %v", er.RoomID)
		} else {
			log.Printf("[GetExamRooms] Room found: %v", room)
		}
		var studentListObjs []interface{}
		for _, studentListID := range er.StudentListIDs {
			studentList, _ := h.service.repo.FindStudentListByID(context.Background(), studentListID)
			if studentList == nil {
				log.Printf("[GetExamRooms] StudentList not found for ID: %v", studentListID)
			} else {
				log.Printf("[GetExamRooms] StudentList found: %v", studentList)
				studentListObjs = append(studentListObjs, studentList)
			}
		}
		var invigilatorObjs []interface{}
		for _, invID := range er.Invigilators {
			inv, _ := h.service.repo.FindUserByID(context.Background(), invID)
			if inv == nil {
				log.Printf("[GetExamRooms] Invigilator not found for ID: %v", invID)
			} else {
				log.Printf("[GetExamRooms] Invigilator found: %v", inv)
				invigilatorObjs = append(invigilatorObjs, inv)
			}
		}
		result = append(result, map[string]interface{}{
			"_id":              er.ID,
			"room":             room,
			"student_lists":    studentListObjs,
			"invigilators":     invigilatorObjs,
			"student_list_ids": er.StudentListIDs,
			"invigilator_ids":  er.Invigilators,
		})
	}

	return c.JSON(http.StatusOK, result)
}

// GetMySeatingPlans returns seating plans where the logged-in student is assigned a seat (by StudentID/CMSID)
func (h *SeatingHandler) GetMySeatingPlans(c echo.Context) error {
	claims, ok := c.Get("user").(*auth.JWTClaims)
	if !ok || claims.CMSID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized or missing StudentID"})
	}
	studentID := claims.CMSID

	plans, err := h.service.GetSeatingPlansByStudentID(c.Request().Context(), studentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch plans"})
	}
	if plans == nil {
		plans = []*SeatingPlan{} // Return empty array if no plans found
	}
	return c.JSON(http.StatusOK, plans)
}

// GetStudentListsByFaculty returns all student lists for the admin's faculty
func (h *SeatingHandler) GetStudentListsByFaculty(c echo.Context) error {
	claims, ok := c.Get("user").(*auth.JWTClaims)
	if !ok || claims.Faculty == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Faculty not found in token"})
	}
	faculty := claims.Faculty
	lists, err := h.service.repo.ListStudentListsByFaculty(c.Request().Context(), faculty)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch student lists"})
	}
	return c.JSON(http.StatusOK, lists)
}

// DeleteSeatingPlan allows admins to delete a seating plan by ID.
func (h *SeatingHandler) DeleteSeatingPlan(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing seating plan ID"})
	}
	planID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid seating plan ID"})
	}
	err = h.service.DeleteSeatingPlan(c.Request().Context(), planID)
	if err != nil {
		if err.Error() == "seating plan not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Seating plan not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete seating plan: " + err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Seating plan deleted successfully"})
}

// DeleteRoom allows admins to delete a room by ID.
func (h *SeatingHandler) DeleteRoom(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing room ID"})
	}
	roomID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid room ID"})
	}
	err = h.service.DeleteRoom(c.Request().Context(), roomID)
	if err != nil {
		if err.Error() == "room not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Room not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete room"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Room deleted successfully"})
}

// ClearRoomAssignments removes all room assignments for a specific exam.
func (h *SeatingHandler) ClearRoomAssignments(c echo.Context) error {
	examID := c.Param("examId")
	if examID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Exam ID is required"})
	}

	id, err := primitive.ObjectIDFromHex(examID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid exam ID"})
	}

	err = h.service.repo.ClearRoomAssignments(context.Background(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to clear room assignments"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Room assignments cleared successfully"})
}

// Why: This handler provides HTTP interfaces for all seating-related operations, with proper validation and error handling for each endpoint.
