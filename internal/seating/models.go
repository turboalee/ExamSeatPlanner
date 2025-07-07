package seating

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Student represents a student in the seating system.
type Student struct {
	StudentID string `bson:"student_id" json:"student_id"`
	Name      string `bson:"name" json:"name"`
}

// StudentList represents a batch of students uploaded together
type StudentList struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Department string             `bson:"department" json:"department"`
	Batch      string             `bson:"batch" json:"batch"`
	Faculty    string             `bson:"faculty" json:"faculty"`
	Name       string             `bson:"name" json:"name"`
	Students   []Student          `bson:"students" json:"students"`
	UploadedBy string             `bson:"uploaded_by" json:"uploaded_by"`
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
	ID        primitive.ObjectID `bson:"_id,omitempty"` // Unique identifier for the exam
	Title     string             `bson:"title"`         // Exam title/course name
	Date      time.Time          `bson:"date"`          // Exam date and time
	Duration  int                `bson:"duration"`      // Exam duration in minutes
	Faculty   string             `bson:"faculty"`       // Faculty conducting the exam
	Algorithm string             `bson:"algorithm"`     // Preferred seating algorithm (matrix, parallel, random)
	CreatedAt time.Time          `bson:"created_at"`    // When the exam was created
	UpdatedAt time.Time          `bson:"updated_at"`    // When the exam was last updated
}

// ExamRoom represents a room assigned to an exam with its students and invigilators
type ExamRoom struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty"`    // Unique identifier for the exam room
	ExamID         primitive.ObjectID   `bson:"exam_id"`          // Reference to the exam
	RoomID         primitive.ObjectID   `bson:"room_id"`          // Reference to the room
	StudentListIDs []primitive.ObjectID `bson:"student_list_ids"` // References to the student lists assigned to this room
	Invigilators   []primitive.ObjectID `bson:"invigilators"`     // List of invigilator IDs assigned to this room
	CreatedAt      time.Time            `bson:"created_at"`       // When the room was assigned
	UpdatedAt      time.Time            `bson:"updated_at"`       // When the room was last updated
}

// UserBasicInfo is a minimal user struct for embedding in plans
// (new struct)
type UserBasicInfo struct {
	ID   primitive.ObjectID `bson:"_id" json:"_id"`
	Name string             `bson:"name" json:"name"`
}

// SeatingPlanRoom represents a room's seating and invigilator assignments within a plan
// (new struct)
type SeatingPlanRoom struct {
	RoomID             primitive.ObjectID   `bson:"room_id" json:"room_id"`
	Name               string               `bson:"name" json:"name"`
	Building           string               `bson:"building" json:"building"`
	Capacity           int                  `bson:"capacity" json:"capacity"`
	Rows               int                  `bson:"rows" json:"rows"`
	Columns            int                  `bson:"columns" json:"columns"`
	Invigilators       []primitive.ObjectID `bson:"invigilators" json:"invigilators"`
	InvigilatorDetails []UserBasicInfo      `bson:"invigilator_details" json:"invigilatorDetails"`
	Seats              []Seat               `bson:"seats" json:"seats"`
}

// SeatingPlan represents a seating arrangement for an exam (now includes all rooms)
type SeatingPlan struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	ExamID    primitive.ObjectID `bson:"exam_id" json:"exam_id"`
	Algorithm string             `bson:"algorithm" json:"algorithm"`
	Status    string             `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	Rooms     []SeatingPlanRoom  `bson:"rooms" json:"rooms"`
}

// Seat represents a single seat assignment in a seating plan.
type Seat struct {
	Row       int    `bson:"row"`        // Row number (1-based)
	Column    int    `bson:"column"`     // Column number (1-based)
	StudentID string `bson:"student_id"` // Student ID (string)
	IsEmpty   bool   `bson:"is_empty"`   // Whether the seat is empty
}

// Why: These models provide the complete data structure for managing exams, rooms, students, invigilators, and seating arrangements with proper relationships and metadata.
