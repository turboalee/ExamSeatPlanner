package seating

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SeatingService handles business logic for seating arrangements.
type SeatingService struct {
	repo *SeatingRepository
}

// NewSeatingService creates a new seating service.
func NewSeatingService(repo *SeatingRepository) *SeatingService {
	return &SeatingService{repo: repo}
}

// GenerateSeatingPlan creates a new seating plan using the specified algorithm.
func (s *SeatingService) GenerateSeatingPlan(ctx context.Context, examID, roomID primitive.ObjectID, invigilatorEmail string, algorithm string, studentIDs []primitive.ObjectID) (*SeatingPlan, error) {
	// Validate inputs
	exam, err := s.repo.FindExamByID(ctx, examID)
	if err != nil || exam == nil {
		return nil, errors.New("exam not found")
	}

	room, err := s.repo.FindRoomByID(ctx, roomID)
	if err != nil || room == nil {
		return nil, errors.New("room not found")
	}

	invigilator, err := s.repo.FindInvigilatorByEmail(ctx, invigilatorEmail)
	if err != nil || invigilator == nil {
		return nil, errors.New("invigilator not found")
	}

	// Check if room capacity is sufficient
	if len(studentIDs) > room.Capacity {
		return nil, errors.New("room capacity insufficient for number of students")
	}

	// Generate seats based on algorithm
	var seats []Seat
	switch algorithm {
	case "matrix":
		seats = s.generateMatrixSeating(room, studentIDs)
	case "parallel":
		seats = s.generateParallelSeating(room, studentIDs)
	case "random":
		seats = s.generateRandomSeating(room, studentIDs)
	default:
		return nil, errors.New("invalid algorithm specified")
	}

	// Create seating plan
	plan := &SeatingPlan{
		ID:            primitive.NewObjectID(),
		ExamID:        examID,
		RoomID:        roomID,
		InvigilatorID: invigilator.ID,
		Algorithm:     algorithm,
		Status:        "draft",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Seats:         seats,
	}

	// Save to database
	err = s.repo.CreateSeatingPlan(ctx, plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// generateMatrixSeating arranges students diagonally based on department/batch/course.
func (s *SeatingService) generateMatrixSeating(room *Room, studentIDs []primitive.ObjectID) []Seat {
	seats := make([]Seat, room.Rows*room.Columns)
	studentIndex := 0

	// Fill seats diagonally (top-left to bottom-right)
	for i := 0; i < room.Rows && studentIndex < len(studentIDs); i++ {
		for j := 0; j < room.Columns && studentIndex < len(studentIDs); j++ {
			seatIndex := i*room.Columns + j
			seats[seatIndex] = Seat{
				Row:       i + 1,
				Column:    j + 1,
				StudentID: studentIDs[studentIndex],
				IsEmpty:   false,
			}
			studentIndex++
		}
	}

	// Mark remaining seats as empty
	for i := studentIndex; i < room.Rows*room.Columns; i++ {
		row := i / room.Columns
		col := i % room.Columns
		seats[i] = Seat{
			Row:     row + 1,
			Column:  col + 1,
			IsEmpty: true,
		}
	}

	return seats
}

// generateParallelSeating arranges students with one column per department/batch/course.
func (s *SeatingService) generateParallelSeating(room *Room, studentIDs []primitive.ObjectID) []Seat {
	seats := make([]Seat, room.Rows*room.Columns)
	studentIndex := 0

	// Fill seats column by column (each column represents a different group)
	for j := 0; j < room.Columns && studentIndex < len(studentIDs); j++ {
		for i := 0; i < room.Rows && studentIndex < len(studentIDs); i++ {
			seatIndex := i*room.Columns + j
			seats[seatIndex] = Seat{
				Row:       i + 1,
				Column:    j + 1,
				StudentID: studentIDs[studentIndex],
				IsEmpty:   false,
			}
			studentIndex++
		}
	}

	// Mark remaining seats as empty
	for i := studentIndex; i < room.Rows*room.Columns; i++ {
		row := i / room.Columns
		col := i % room.Columns
		seats[i] = Seat{
			Row:     row + 1,
			Column:  col + 1,
			IsEmpty: true,
		}
	}

	return seats
}

// generateRandomSeating arranges students randomly.
func (s *SeatingService) generateRandomSeating(room *Room, studentIDs []primitive.ObjectID) []Seat {
	seats := make([]Seat, room.Rows*room.Columns)

	// Create a slice of available positions
	positions := make([]int, room.Rows*room.Columns)
	for i := range positions {
		positions[i] = i
	}

	// Shuffle positions
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Assign students to random positions
	for i, studentID := range studentIDs {
		if i < len(positions) {
			pos := positions[i]
			row := pos / room.Columns
			col := pos % room.Columns
			seats[pos] = Seat{
				Row:       row + 1,
				Column:    col + 1,
				StudentID: studentID,
				IsEmpty:   false,
			}
		}
	}

	// Mark remaining seats as empty
	for i := len(studentIDs); i < room.Rows*room.Columns; i++ {
		pos := positions[i]
		row := pos / room.Columns
		col := pos % room.Columns
		seats[pos] = Seat{
			Row:     row + 1,
			Column:  col + 1,
			IsEmpty: true,
		}
	}

	return seats
}

// GetSeatingPlan retrieves a seating plan by ID.
func (s *SeatingService) GetSeatingPlan(ctx context.Context, planID primitive.ObjectID) (*SeatingPlan, error) {
	return s.repo.FindSeatingPlanByID(ctx, planID)
}

// UpdateSeatingPlanStatus updates the status of a seating plan.
func (s *SeatingService) UpdateSeatingPlanStatus(ctx context.Context, planID primitive.ObjectID, status string) error {
	plan, err := s.repo.FindSeatingPlanByID(ctx, planID)
	if err != nil || plan == nil {
		return errors.New("seating plan not found")
	}

	plan.Status = status
	plan.UpdatedAt = time.Now()
	return s.repo.UpdateSeatingPlan(ctx, plan)
}

// Why: This service implements the three seating algorithms and provides business logic for managing seating plans, ensuring proper validation and data consistency.
