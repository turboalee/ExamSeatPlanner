package seating

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Student represents a student in the seating system.
type Student struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"` // Unique identifier for the student
	StudentID  string             `bson:"student_id"`    // Student's ID for identification
	Name       string             `bson:"name"`          // Student's full name
	Email      string             `bson:"email"`         // Student's email for notifications
	Department string             `bson:"department"`    // Student's department for grouping
	Batch      string             `bson:"batch"`         // Student's batch/year for grouping
	Course     string             `bson:"course"`        // Student's course for grouping
	Faculty    string             `bson:"faculty"`       // Student's faculty for grouping
}

// Room represents an examination room.
type Room struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"` // Unique identifier for the room
	Name     string             `bson:"name"`          // Room name/number
	Capacity int                `bson:"capacity"`      // Total number of seats (rows * columns)
	Rows     int                `bson:"rows"`          // Number of rows in the room
	Columns  int                `bson:"columns"`       // Number of columns in the room
	Building string             `bson:"building"`      // Building where room is located
}

// Invigilator represents an exam invigilator.
type Invigilator struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"` // Unique identifier for the invigilator
	Email   string             `bson:"email"`         // Invigilator's email (primary identifier)
	Name    string             `bson:"name"`          // Invigilator's full name
	Faculty string             `bson:"faculty"`       // Invigilator's faculty
}

// Exam represents an examination event.
type Exam struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`  // Unique identifier for the exam
	Title         string             `bson:"title"`          // Exam title/course name
	Date          time.Time          `bson:"date"`           // Exam date and time
	Duration      int                `bson:"duration"`       // Exam duration in minutes
	Faculty       string             `bson:"faculty"`        // Faculty conducting the exam
	Department    string             `bson:"department"`     // Department conducting the exam
	Course        string             `bson:"course"`         // Course code
	Batch         string             `bson:"batch"`          // Batch taking the exam
	TotalStudents int                `bson:"total_students"` // Total number of students
}

// SeatingPlan represents a seating arrangement for an exam.
type SeatingPlan struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`  // Unique identifier for the seating plan
	ExamID        primitive.ObjectID `bson:"exam_id"`        // Reference to the exam
	RoomID        primitive.ObjectID `bson:"room_id"`        // Reference to the room
	InvigilatorID primitive.ObjectID `bson:"invigilator_id"` // Reference to the invigilator
	Algorithm     string             `bson:"algorithm"`      // Algorithm used (matrix, parallel, random)
	Status        string             `bson:"status"`         // Status (draft, final, published)
	CreatedAt     time.Time          `bson:"created_at"`     // When the plan was created
	UpdatedAt     time.Time          `bson:"updated_at"`     // When the plan was last updated
	Seats         []Seat             `bson:"seats"`          // Array of seat assignments
}

// Seat represents a single seat assignment in a seating plan.
type Seat struct {
	Row       int                `bson:"row"`        // Row number (1-based)
	Column    int                `bson:"column"`     // Column number (1-based)
	StudentID primitive.ObjectID `bson:"student_id"` // Reference to the student
	IsEmpty   bool               `bson:"is_empty"`   // Whether the seat is empty
}

// Why: These models provide the complete data structure for managing exams, rooms, students, invigilators, and seating arrangements with proper relationships and metadata.
