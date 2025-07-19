package seating

import (
	"context"
	"errors"
	"fmt" // Added for debug printing
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
func (s *SeatingService) GenerateSeatingPlan(ctx context.Context, examID, _ primitive.ObjectID, invigilatorEmail string, algorithm string, _ []primitive.ObjectID) ([]*SeatingPlan, error) {
	// 1. Fetch exam
	exam, err := s.repo.FindExamByID(ctx, examID)
	if err != nil || exam == nil {
		return nil, errors.New("exam not found")
	}

	// 2. Fetch exam rooms for this exam
	examRooms, err := s.repo.GetExamRooms(ctx, examID)
	if err != nil || len(examRooms) == 0 {
		return nil, errors.New("no rooms assigned to this exam")
	}

	var allRooms []*Room
	var roomExamRooms []*ExamRoom
	var roomStudentsList [][]StudentWithGroup
	assignedStudentIDs := make(map[string]bool)

	for _, examRoom := range examRooms {
		// Fetch room details
		room, err := s.repo.FindRoomByID(ctx, examRoom.RoomID)
		if err != nil || room == nil {
			continue // Skip invalid rooms
		}
		allRooms = append(allRooms, room)
		roomExamRooms = append(roomExamRooms, examRoom)

		// Fetch all student lists for this room
		studentLists, err := s.repo.FindStudentListsByIDs(ctx, examRoom.StudentListIDs)
		if err != nil || len(studentLists) == 0 {
			roomStudentsList = append(roomStudentsList, []StudentWithGroup{})
			continue
		}

		// Gather unassigned students from all lists for this room
		var studentsForRoom []StudentWithGroup
		for _, list := range studentLists {
			for _, student := range list.Students {
				if student.StudentID != "" {
					studentsForRoom = append(studentsForRoom, StudentWithGroup{
						StudentID:  student.StudentID,
						Name:       student.Name,
						Department: list.Department,
						Batch:      list.Batch,
					})
				}
			}
		}
		// Debug log: print all students being assigned to this room
		var ids []string
		for _, s := range studentsForRoom {
			ids = append(ids, s.StudentID)
		}
		fmt.Printf("[DEBUG] StudentIDs for room %s: %+v\n", room.Name, ids)
		// Only assign up to room capacity
		if len(studentsForRoom) > room.Capacity {
			studentsForRoom = studentsForRoom[:room.Capacity]
		}
		// Mark these students as assigned
		for _, s := range studentsForRoom {
			assignedStudentIDs[s.StudentID] = true
		}
		roomStudentsList = append(roomStudentsList, studentsForRoom)
	}

	// 4. Calculate total capacity
	totalCapacity := 0
	for _, room := range allRooms {
		totalCapacity += room.Capacity
	}

	totalStudents := 0
	for _, students := range roomStudentsList {
		totalStudents += len(students)
	}

	if totalStudents > totalCapacity {
		return nil, errors.New("total students exceed total room capacity")
	}

	// 5. Build the plan with all rooms, applying the algorithm per room
	planRooms := make([]SeatingPlanRoom, 0)
	for i, room := range allRooms {
		examRoom := roomExamRooms[i]

		// Fetch invigilator details
		var invigilatorDetails []UserBasicInfo
		for _, invID := range examRoom.Invigilators {
			user, err := s.repo.FindUserByID(ctx, invID)
			if err == nil && user != nil {
				invigilatorDetails = append(invigilatorDetails, UserBasicInfo{
					ID:   user.ID,
					Name: user.Name,
				})
			}
		}

		roomStudents := roomStudentsList[i]
		var seats []Seat

		if len(roomStudents) > 0 {
			// Generate seats for this room using the specified algorithm
			switch algorithm {
			case "parallel":
				seats = s.generateParallelSeating(room, roomStudents)
			case "simple":
				seats = s.generateRandomSeating(room, roomStudents)
			case "separated":
				var err error
				seats, err = s.generateSnakeSeating(room, roomStudents)
				if err != nil {
					return nil, err
				}
			default:
				return nil, errors.New("invalid algorithm specified: must be 'parallel', 'simple', or 'separated'")
			}
		} else {
			// Create empty seats for this room
			seats = make([]Seat, room.Rows*room.Columns)
			for i := 0; i < room.Rows*room.Columns; i++ {
				row := i / room.Columns
				col := i % room.Columns
				seats[i] = Seat{
					Row:     row + 1,
					Column:  col + 1,
					IsEmpty: true,
				}
			}
		}

		planRoom := SeatingPlanRoom{
			RoomID:             room.ID,
			Name:               room.Name,
			Building:           room.Building,
			Capacity:           room.Capacity,
			Rows:               room.Rows,
			Columns:            room.Columns,
			Invigilators:       examRoom.Invigilators,
			InvigilatorDetails: invigilatorDetails,
			Seats:              seats,
		}
		planRooms = append(planRooms, planRoom)
	}

	// Defensive: ensure Rooms is always a non-nil slice
	if planRooms == nil {
		planRooms = []SeatingPlanRoom{}
	}

	plan := &SeatingPlan{
		ID:        primitive.NewObjectID(),
		ExamID:    examID,
		Algorithm: algorithm,
		Status:    "draft",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Rooms:     planRooms,
	}
	err = s.repo.CreateSeatingPlan(ctx, plan)
	if err != nil {
		return nil, err
	}

	return []*SeatingPlan{plan}, nil
}

// distributeStudentsAcrossRooms distributes students sequentially across rooms, filling each room up to its capacity.
func (s *SeatingService) distributeStudentsAcrossRooms(allStudents []StudentWithGroup, rooms []*Room, algorithm string) [][]StudentWithGroup {
	fmt.Printf("[DEBUG] Algorithm: %s\n", algorithm)
	fmt.Printf("[DEBUG] Total students to distribute: %d\n", len(allStudents))
	deptCount := map[string]int{}
	for _, s := range allStudents {
		deptCount[s.Department]++
	}
	fmt.Printf("[DEBUG] Students per department: %v\n", deptCount)
	result := make([][]StudentWithGroup, len(rooms))
	for i := range result {
		result[i] = make([]StudentWithGroup, 0)
	}

	switch algorithm {
	case "matrix":
		// Group students by department
		deptMap := map[string][]StudentWithGroup{}
		var depts []string
		for _, s := range allStudents {
			if _, ok := deptMap[s.Department]; !ok {
				depts = append(depts, s.Department)
			}
			deptMap[s.Department] = append(deptMap[s.Department], s)
		}
		// For each room, assign as even a split as possible
		for roomIdx, room := range rooms {
			cap := room.Capacity
			totalLeft := 0
			for _, d := range depts {
				totalLeft += len(deptMap[d])
			}
			if totalLeft == 0 {
				continue
			}
			// Proportional allocation
			alloc := make(map[string]int)
			left := cap
			for i, d := range depts {
				if i == len(depts)-1 {
					alloc[d] = left // assign the rest to the last dept
				} else {
					want := (len(deptMap[d]) * cap) / totalLeft
					if want > len(deptMap[d]) {
						want = len(deptMap[d])
					}
					alloc[d] = want
					left -= want
				}
			}
			// Assign students to this room
			for _, d := range depts {
				count := alloc[d]
				for i := 0; i < count && len(deptMap[d]) > 0; i++ {
					result[roomIdx] = append(result[roomIdx], deptMap[d][0])
					deptMap[d] = deptMap[d][1:]
				}
			}
			// Fill any remaining seats round-robin from remaining students
			deptIdx := 0
			for len(result[roomIdx]) < cap {
				found := false
				for tries := 0; tries < len(depts); tries++ {
					d := depts[deptIdx%len(depts)]
					if len(deptMap[d]) > 0 {
						result[roomIdx] = append(result[roomIdx], deptMap[d][0])
						deptMap[d] = deptMap[d][1:]
						found = true
						break
					}
					deptIdx++
				}
				if !found {
					break // no more students left
				}
			}
		}
	case "random":
		// Shuffle all students
		students := make([]StudentWithGroup, len(allStudents))
		copy(students, allStudents)
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(students), func(i, j int) { students[i], students[j] = students[j], students[i] })
		// Assign to rooms in round-robin order
		roomIdx := 0
		for _, s := range students {
			for result[roomIdx] != nil && len(result[roomIdx]) >= rooms[roomIdx].Capacity {
				roomIdx = (roomIdx + 1) % len(rooms)
			}
			result[roomIdx] = append(result[roomIdx], s)
			roomIdx = (roomIdx + 1) % len(rooms)
		}
	case "parallel":
		// Fill each room with as much of a department as possible before moving to the next
		deptMap := map[string][]StudentWithGroup{}
		var depts []string
		for _, s := range allStudents {
			if _, ok := deptMap[s.Department]; !ok {
				depts = append(depts, s.Department)
			}
			deptMap[s.Department] = append(deptMap[s.Department], s)
		}
		roomIdx := 0
		for _, dept := range depts {
			students := deptMap[dept]
			idx := 0
			for idx < len(students) {
				capLeft := rooms[roomIdx].Capacity - len(result[roomIdx])
				toAssign := min(capLeft, len(students)-idx)
				result[roomIdx] = append(result[roomIdx], students[idx:idx+toAssign]...)
				idx += toAssign
				if len(result[roomIdx]) >= rooms[roomIdx].Capacity {
					roomIdx++
					if roomIdx >= len(rooms) {
						break
					}
				}
			}
		}
	default:
		// Fallback: sequential fill
		idx := 0
		for _, s := range allStudents {
			for result[idx] != nil && len(result[idx]) >= rooms[idx].Capacity {
				idx = (idx + 1) % len(rooms)
			}
			result[idx] = append(result[idx], s)
			idx = (idx + 1) % len(rooms)
		}
	}

	fmt.Println("[DEBUG] Department composition per room:")
	for i, roomStudents := range result {
		deptCount := map[string]int{}
		for _, s := range roomStudents {
			deptCount[s.Department]++
		}
		fmt.Printf("Room %d: %v\n", i+1, deptCount)
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type StudentWithGroup struct {
	StudentID  string
	Name       string
	Department string
	Batch      string
}

// generateParallelSeating arranges students by department per column.
func (s *SeatingService) generateParallelSeating(room *Room, students []StudentWithGroup) []Seat {
	fmt.Printf("[DEBUG] generateParallelSeating CALLED for room: %s with %d students\n", room.Name, len(students))
	seats := make([]Seat, room.Rows*room.Columns)
	// Group students by department
	deptMap := map[string][]StudentWithGroup{}
	var depts []string
	for _, student := range students {
		if _, ok := deptMap[student.Department]; !ok {
			depts = append(depts, student.Department)
		}
		deptMap[student.Department] = append(deptMap[student.Department], student)
	}
	// Assign each department to a column (cycle if more columns than depts)
	studentIndex := 0
	colDept := make([]string, room.Columns)
	for i := 0; i < room.Columns; i++ {
		colDept[i] = depts[i%len(depts)]
	}
	// For each column, fill with students from the assigned department
	colStudentIdx := make(map[string]int)
	for j := 0; j < room.Columns; j++ {
		dept := colDept[j]
		for i := 0; i < room.Rows; i++ {
			seatIndex := i*room.Columns + j
			idx := colStudentIdx[dept]
			if idx < len(deptMap[dept]) {
				s := deptMap[dept][idx]
				seats[seatIndex] = Seat{
					Row:       i + 1,
					Column:    j + 1,
					StudentID: s.StudentID, // Always set StudentID
					IsEmpty:   false,
				}
				colStudentIdx[dept]++
				studentIndex++
			} else {
				seats[seatIndex] = Seat{
					Row:       i + 1,
					Column:    j + 1,
					StudentID: "", // Explicitly set to empty string
					IsEmpty:   true,
				}
			}
		}
	}
	// Debug log
	fmt.Printf("[DEBUG] generateParallelSeating: first 5 seats: %+v\n", seats[:min(5, len(seats))])
	var studentIDs []string
	for i := 0; i < min(5, len(seats)); i++ {
		studentIDs = append(studentIDs, seats[i].StudentID)
	}
	fmt.Printf("[DEBUG] generateParallelSeating: first 5 seat StudentIDs: %+v\n", studentIDs)
	// Debug log
	fmt.Printf("[DEBUG] generateParallelSeating: ALL seat StudentIDs for room %s: %+v\n", room.Name, studentIDs)
	return seats
}

// generateRandomSeating arranges students in a classic snake/serpentine (row-wise, alternating direction) order, interleaving departments in round-robin order, with no adjacency constraints.
func (s *SeatingService) generateRandomSeating(room *Room, students []StudentWithGroup) []Seat {
	fmt.Printf("[DEBUG] generateRandomSeating (classic snake/serpentine, round-robin interleaving) CALLED for room: %s with %d students\n", room.Name, len(students))
	seats := make([]Seat, room.Rows*room.Columns)
	// Group students by department
	deptMap := map[string][]StudentWithGroup{}
	var depts []string
	for _, s := range students {
		if _, ok := deptMap[s.Department]; !ok {
			depts = append(depts, s.Department)
		}
		deptMap[s.Department] = append(deptMap[s.Department], s)
	}
	studentCount := len(students)
	studentIndex := 0
	deptIdx := 0
	for i := 0; i < room.Rows; i++ {
		if i%2 == 0 { // Even row: left-to-right
			for j := 0; j < room.Columns; j++ {
				seatIdx := i*room.Columns + j
				if studentIndex < studentCount {
					// Find next department with students left
					tries := 0
					for tries < len(depts) {
						dept := depts[deptIdx%len(depts)]
						if len(deptMap[dept]) > 0 {
							s := deptMap[dept][0]
							deptMap[dept] = deptMap[dept][1:]
							seats[seatIdx] = Seat{
								Row:       i + 1,
								Column:    j + 1,
								StudentID: s.StudentID,
								IsEmpty:   false,
							}
							studentIndex++
							deptIdx++
							break
						} else {
							deptIdx++
							tries++
						}
					}
					if tries == len(depts) {
						// No students left in any department
						seats[seatIdx] = Seat{
							Row:     i + 1,
							Column:  j + 1,
							IsEmpty: true,
						}
					}
				} else {
					seats[seatIdx] = Seat{
						Row:     i + 1,
						Column:  j + 1,
						IsEmpty: true,
					}
				}
			}
		} else { // Odd row: right-to-left
			for j := room.Columns - 1; j >= 0; j-- {
				seatIdx := i*room.Columns + j
				if studentIndex < studentCount {
					// Find next department with students left
					tries := 0
					for tries < len(depts) {
						dept := depts[deptIdx%len(depts)]
						if len(deptMap[dept]) > 0 {
							s := deptMap[dept][0]
							deptMap[dept] = deptMap[dept][1:]
							seats[seatIdx] = Seat{
								Row:       i + 1,
								Column:    j + 1,
								StudentID: s.StudentID,
								IsEmpty:   false,
							}
							studentIndex++
							deptIdx++
							break
						} else {
							deptIdx++
							tries++
						}
					}
					if tries == len(depts) {
						// No students left in any department
						seats[seatIdx] = Seat{
							Row:     i + 1,
							Column:  j + 1,
							IsEmpty: true,
						}
					}
				} else {
					seats[seatIdx] = Seat{
						Row:     i + 1,
						Column:  j + 1,
						IsEmpty: true,
					}
				}
			}
		}
	}
	return seats
}

// generateSnakeSeating arranges students to minimize same-department adjacency in both rows and columns.
func (s *SeatingService) generateSnakeSeating(room *Room, students []StudentWithGroup) ([]Seat, error) {
	fmt.Printf("[DEBUG] generateSnakeSeating (robust empty seats) CALLED for room: %s with %d students\n", room.Name, len(students))
	seats := make([]Seat, room.Rows*room.Columns)
	// Group students by department
	deptMap := map[string][]StudentWithGroup{}
	for _, s := range students {
		deptMap[s.Department] = append(deptMap[s.Department], s)
	}
	// Helper: get department of a student by StudentID
	studentDept := map[string]string{}
	for _, s := range students {
		studentDept[s.StudentID] = s.Department
	}
	for i := 0; i < room.Rows; i++ {
		for j := 0; j < room.Columns; j++ {
			seatIdx := i*room.Columns + j
			// Check adjacent seats (above and left)
			adjDepts := map[string]bool{}
			if i > 0 {
				above := seats[(i-1)*room.Columns+j]
				if above.StudentID != "" {
					if dept, ok := studentDept[above.StudentID]; ok {
						adjDepts[dept] = true
					}
				}
			}
			if j > 0 {
				left := seats[i*room.Columns+(j-1)]
				if left.StudentID != "" {
					if dept, ok := studentDept[left.StudentID]; ok {
						adjDepts[dept] = true
					}
				}
			}
			// Find all departments with students left that are NOT adjacent
			candidates := []string{}
			for dept, group := range deptMap {
				if len(group) > 0 && !adjDepts[dept] {
					candidates = append(candidates, dept)
				}
			}
			if len(candidates) == 0 {
				// No valid department, leave seat empty
				seats[seatIdx] = Seat{Row: i + 1, Column: j + 1, IsEmpty: true}
				continue
			}
			// Pick the first available department
			dept := candidates[0]
			s := deptMap[dept][0]
			deptMap[dept] = deptMap[dept][1:]
			seats[seatIdx] = Seat{
				Row:       i + 1,
				Column:    j + 1,
				StudentID: s.StudentID,
				IsEmpty:   false,
			}
		}
	}
	// After assignment, check if any students remain unassigned
	unassigned := 0
	for _, group := range deptMap {
		unassigned += len(group)
	}
	if unassigned > 0 {
		return nil, fmt.Errorf("Not all students can be accommodated with the current constraints. Unassigned students: %d", unassigned)
	}
	return seats, nil
}

// GetSeatingPlan retrieves a seating plan by ID.
func (s *SeatingService) GetSeatingPlan(ctx context.Context, planID primitive.ObjectID) (*SeatingPlan, error) {
	return s.repo.FindSeatingPlanByID(ctx, planID)
}

// UpdateSeatingPlanStatus updates the status of a seating plan.
func (s *SeatingService) UpdateSeatingPlanStatus(ctx context.Context, planID primitive.ObjectID, status string) error {
	plan, err := s.repo.FindSeatingPlanByID(ctx, planID)
	if err != nil {
		return err
	}
	if plan == nil {
		return errors.New("seating plan not found")
	}

	plan.Status = status
	plan.UpdatedAt = time.Now()
	return s.repo.UpdateSeatingPlan(ctx, plan)
}

// DeleteSeatingPlan deletes a seating plan by ID.
func (s *SeatingService) DeleteSeatingPlan(ctx context.Context, planID primitive.ObjectID) error {
	return s.repo.DeleteSeatingPlan(ctx, planID)
}

// GetAllExams retrieves all exams.
func (s *SeatingService) GetAllExams(ctx context.Context) ([]*Exam, error) {
	return s.repo.GetAllExams(ctx)
}

// GetAllStudents retrieves all students.
func (s *SeatingService) GetAllStudents(ctx context.Context) ([]*Student, error) {
	return s.repo.GetAllStudents(ctx)
}

// GetAllSeatingPlans retrieves all seating plans.
func (s *SeatingService) GetAllSeatingPlans(ctx context.Context) ([]*SeatingPlan, error) {
	return s.repo.GetAllSeatingPlans(ctx)
}

// GetAllRooms retrieves all rooms.
func (s *SeatingService) GetAllRooms(ctx context.Context) ([]*Room, error) {
	return s.repo.GetAllRooms(ctx)
}

// GetAllStudentLists retrieves all student lists.
func (s *SeatingService) GetAllStudentLists(ctx context.Context) ([]*StudentList, error) {
	return s.repo.GetAllStudentLists(ctx)
}

// GetAllInvigilators retrieves all invigilators (now users with role admin or staff)
func (s *SeatingService) GetAllInvigilators(ctx context.Context) ([]*User, error) {
	return s.repo.GetAllInvigilators(ctx)
}

// GetExamRooms retrieves all rooms for a specific exam.
func (s *SeatingService) GetExamRooms(ctx context.Context, examID primitive.ObjectID) ([]*ExamRoom, error) {
	return s.repo.GetExamRooms(ctx, examID)
}

// GetSeatingPlansByStudentID returns seating plans where a seat.student_id matches the given StudentID
func (s *SeatingService) GetSeatingPlansByStudentID(ctx context.Context, studentID string) ([]*SeatingPlan, error) {
	return s.repo.FindSeatingPlansByStudentID(ctx, studentID)
}

func (s *SeatingService) DeleteRoom(ctx context.Context, roomID primitive.ObjectID) error {
	return s.repo.DeleteRoom(ctx, roomID)
}

func (s *SeatingService) UpdateRoom(ctx context.Context, roomID primitive.ObjectID, room *Room) error {
	return s.repo.UpdateRoom(ctx, roomID, room)
}
