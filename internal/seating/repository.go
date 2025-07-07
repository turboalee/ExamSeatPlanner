package seating

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// User struct for invigilator queries (copied from internal/auth/models.go)
type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	CMSID      string             `bson:"cms_id" json:"cms_id"`
	Name       string             `bson:"name" json:"name"`
	Email      string             `bson:"email" json:"email"`
	Role       string             `bson:"role" json:"role"`
	Faculty    string             `bson:"faculty" json:"faculty"`
	Department string             `bson:"department" json:"department"`
	Batch      string             `bson:"batch" json:"batch"`
}

// SeatingRepository handles DB operations for seating-related entities.
type SeatingRepository struct {
	studentsCollection     *mongo.Collection
	roomsCollection        *mongo.Collection
	examsCollection        *mongo.Collection
	invigilatorsCollection *mongo.Collection
	seatingPlansCollection *mongo.Collection
	studentListsCollection *mongo.Collection
	examRoomsCollection    *mongo.Collection
	usersCollection        *mongo.Collection
}

// NewSeatingRepository creates a new repository for seating operations.
func NewSeatingRepository(db *mongo.Database) *SeatingRepository {
	return &SeatingRepository{
		studentsCollection:     db.Collection("students"),
		roomsCollection:        db.Collection("rooms"),
		examsCollection:        db.Collection("exams"),
		invigilatorsCollection: db.Collection("invigilators"),
		seatingPlansCollection: db.Collection("seating_plans"),
		studentListsCollection: db.Collection("student_lists"),
		examRoomsCollection:    db.Collection("exam_rooms"),
		usersCollection:        db.Collection("users"),
	}
}

// Student operations
func (r *SeatingRepository) CreateStudent(ctx context.Context, student *Student) error {
	_, err := r.studentsCollection.InsertOne(ctx, student)
	return err
}

func (r *SeatingRepository) FindStudentByID(ctx context.Context, studentID string) (*Student, error) {
	var student Student
	err := r.studentsCollection.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&student)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &student, nil
}

func (r *SeatingRepository) FindStudentsByDepartmentAndBatch(ctx context.Context, department, batch string) ([]*Student, error) {
	filter := bson.M{"department": department, "batch": batch}
	cursor, err := r.studentsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var students []*Student
	if err := cursor.All(ctx, &students); err != nil {
		return nil, err
	}
	return students, nil
}

// Room operations
func (r *SeatingRepository) CreateRoom(ctx context.Context, room *Room) error {
	_, err := r.roomsCollection.InsertOne(ctx, room)
	return err
}

func (r *SeatingRepository) FindRoomByID(ctx context.Context, id primitive.ObjectID) (*Room, error) {
	var room Room
	err := r.roomsCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

func (r *SeatingRepository) FindAllRooms(ctx context.Context) ([]*Room, error) {
	cursor, err := r.roomsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var rooms []*Room
	if err := cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *SeatingRepository) DeleteRoom(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.roomsCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("room not found")
	}
	return nil
}

func (r *SeatingRepository) UpdateRoom(ctx context.Context, id primitive.ObjectID, room *Room) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"name":     room.Name,
			"rows":     room.Rows,
			"columns":  room.Columns,
			"building": room.Building,
			"capacity": room.Capacity,
		},
	}
	res, err := r.roomsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("room not found")
	}
	return nil
}

// Exam operations
func (r *SeatingRepository) CreateExam(ctx context.Context, exam *Exam) error {
	_, err := r.examsCollection.InsertOne(ctx, exam)
	return err
}

func (r *SeatingRepository) FindExamByID(ctx context.Context, id primitive.ObjectID) (*Exam, error) {
	var exam Exam
	err := r.examsCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&exam)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &exam, nil
}

func (r *SeatingRepository) FindExamsByFaculty(ctx context.Context, faculty string) ([]*Exam, error) {
	filter := bson.M{"faculty": faculty}
	cursor, err := r.examsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var exams []*Exam
	if err := cursor.All(ctx, &exams); err != nil {
		return nil, err
	}
	return exams, nil
}

func (r *SeatingRepository) DeleteExam(ctx context.Context, id primitive.ObjectID) error {
	// Delete the exam document
	res, err := r.examsCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("exam not found")
	}
	// Cascade delete: delete all ExamRoom documents for this exam
	_, err = r.examRoomsCollection.DeleteMany(ctx, bson.M{"exam_id": id})
	if err != nil {
		return err
	}
	// Cascade delete: delete all SeatingPlan documents for this exam
	_, err = r.seatingPlansCollection.DeleteMany(ctx, bson.M{"exam_id": id})
	if err != nil {
		return err
	}
	return nil
}

func (r *SeatingRepository) UpdateExam(ctx context.Context, exam *Exam) error {
	filter := bson.M{"_id": exam.ID}
	update := bson.M{"$set": exam}
	res, err := r.examsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("exam not found")
	}
	return nil
}

// Invigilator operations
func (r *SeatingRepository) CreateInvigilator(ctx context.Context, invigilator *Invigilator) error {
	_, err := r.invigilatorsCollection.InsertOne(ctx, invigilator)
	return err
}

func (r *SeatingRepository) FindInvigilatorByEmail(ctx context.Context, email string) (*Invigilator, error) {
	var invigilator Invigilator
	err := r.invigilatorsCollection.FindOne(ctx, bson.M{"email": email}).Decode(&invigilator)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &invigilator, nil
}

// SeatingPlan operations
func (r *SeatingRepository) CreateSeatingPlan(ctx context.Context, plan *SeatingPlan) error {
	_, err := r.seatingPlansCollection.InsertOne(ctx, plan)
	return err
}

func (r *SeatingRepository) FindSeatingPlanByID(ctx context.Context, id primitive.ObjectID) (*SeatingPlan, error) {
	var plan SeatingPlan
	err := r.seatingPlansCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&plan)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

func (r *SeatingRepository) FindSeatingPlansByExam(ctx context.Context, examID primitive.ObjectID) ([]*SeatingPlan, error) {
	filter := bson.M{"exam_id": examID}
	cursor, err := r.seatingPlansCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var plans []*SeatingPlan
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *SeatingRepository) UpdateSeatingPlan(ctx context.Context, plan *SeatingPlan) error {
	filter := bson.M{"_id": plan.ID}
	update := bson.M{"$set": plan}
	res, err := r.seatingPlansCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("seating plan not found")
	}
	return nil
}

// FindSeatingPlansByStudentID returns seating plans where any seat.student_id matches the given StudentID
func (r *SeatingRepository) FindSeatingPlansByStudentID(ctx context.Context, studentID string) ([]*SeatingPlan, error) {
	filter := bson.M{"seats.student_id": studentID}
	cursor, err := r.seatingPlansCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var plans []*SeatingPlan
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, err
	}
	return plans, nil
}

// DeleteSeatingPlan deletes a seating plan by its ID from the seatingPlansCollection.
func (r *SeatingRepository) DeleteSeatingPlan(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.seatingPlansCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("seating plan not found")
	}
	return nil
}

// StudentList operations
// CreateStudentList saves a new student list to the database
func (r *SeatingRepository) CreateStudentList(ctx context.Context, list *StudentList) error {
	_, err := r.studentListsCollection.InsertOne(ctx, list)
	return err
}

func (r *SeatingRepository) FindStudentListByID(ctx context.Context, id primitive.ObjectID) (*StudentList, error) {
	var studentList StudentList
	err := r.studentListsCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&studentList)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &studentList, nil
}

func (r *SeatingRepository) FindAllStudentLists(ctx context.Context) ([]*StudentList, error) {
	cursor, err := r.studentListsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var studentLists []*StudentList
	if err := cursor.All(ctx, &studentLists); err != nil {
		return nil, err
	}
	return studentLists, nil
}

// ListStudentListsByFaculty returns all student lists for a given faculty
func (r *SeatingRepository) ListStudentListsByFaculty(ctx context.Context, faculty string) ([]*StudentList, error) {
	filter := bson.M{"faculty": faculty}
	cursor, err := r.studentListsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var lists []*StudentList
	if err := cursor.All(ctx, &lists); err != nil {
		return nil, err
	}
	return lists, nil
}

// FindStudentListsByIDs fetches multiple student lists by their ObjectIDs
func (r *SeatingRepository) FindStudentListsByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*StudentList, error) {
	if len(ids) == 0 {
		return []*StudentList{}, nil
	}
	filter := bson.M{"_id": bson.M{"$in": ids}}
	cursor, err := r.studentListsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var lists []*StudentList
	if err := cursor.All(ctx, &lists); err != nil {
		return nil, err
	}
	return lists, nil
}

// Add after FindAllStudentLists
func (r *SeatingRepository) DeleteStudentList(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.studentListsCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *SeatingRepository) UpdateStudentList(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	_, err := r.studentListsCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

// Add a student to a student list
func (r *SeatingRepository) AddStudentToList(ctx context.Context, listID primitive.ObjectID, student Student) error {
	update := bson.M{"$addToSet": bson.M{"students": student}}
	_, err := r.studentListsCollection.UpdateOne(ctx, bson.M{"_id": listID}, update)
	return err
}

// Update a student in a student list
func (r *SeatingRepository) UpdateStudentInList(ctx context.Context, listID primitive.ObjectID, studentID string, updated Student) error {
	// Fetch the current student list
	studentList, err := r.FindStudentListByID(ctx, listID)
	if err != nil {
		return err
	}
	if studentList == nil {
		return errors.New("student list not found")
	}
	// Check for duplicate student_id (other than the one being updated)
	for _, s := range studentList.Students {
		if s.StudentID == updated.StudentID && s.StudentID != studentID {
			return errors.New("student_id already exists in this list")
		}
	}
	// Remove the old student by studentID
	pull := bson.M{"$pull": bson.M{"students": bson.M{"student_id": studentID}}}
	res1, err := r.studentListsCollection.UpdateOne(ctx, bson.M{"_id": listID}, pull)
	if err != nil {
		return err
	}
	if res1.ModifiedCount == 0 {
		return errors.New("student not found in list")
	}
	// Add the updated student (with possibly new student_id)
	push := bson.M{"$addToSet": bson.M{"students": updated}}
	_, err = r.studentListsCollection.UpdateOne(ctx, bson.M{"_id": listID}, push)
	return err
}

// Remove a student from a student list
func (r *SeatingRepository) RemoveStudentFromList(ctx context.Context, listID primitive.ObjectID, studentID string) error {
	update := bson.M{"$pull": bson.M{"students": bson.M{"student_id": studentID}}}
	res, err := r.studentListsCollection.UpdateOne(ctx, bson.M{"_id": listID}, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("student not found in list")
	}
	return nil
}

// ExamRoom operations
func (r *SeatingRepository) CreateExamRoom(ctx context.Context, examRoom *ExamRoom) error {
	_, err := r.examRoomsCollection.InsertOne(ctx, examRoom)
	return err
}

func (r *SeatingRepository) FindExamRoomByID(ctx context.Context, id primitive.ObjectID) (*ExamRoom, error) {
	var examRoom ExamRoom
	err := r.examRoomsCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&examRoom)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &examRoom, nil
}

func (r *SeatingRepository) GetExamRooms(ctx context.Context, examID primitive.ObjectID) ([]*ExamRoom, error) {
	filter := bson.M{"exam_id": examID}
	cursor, err := r.examRoomsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var examRooms []*ExamRoom
	if err := cursor.All(ctx, &examRooms); err != nil {
		return nil, err
	}
	return examRooms, nil
}

func (r *SeatingRepository) AddInvigilatorToRoom(ctx context.Context, examRoomID, invigilatorID primitive.ObjectID) error {
	filter := bson.M{"_id": examRoomID}
	update := bson.M{"$addToSet": bson.M{"invigilators": invigilatorID}}
	res, err := r.examRoomsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("exam room not found")
	}
	return nil
}

// ClearRoomAssignments removes all room assignments for a specific exam.
func (r *SeatingRepository) ClearRoomAssignments(ctx context.Context, examID primitive.ObjectID) error {
	collection := r.examRoomsCollection

	// Delete all exam room assignments for the given exam ID
	_, err := collection.DeleteMany(ctx, bson.M{"exam_id": examID})
	if err != nil {
		return err
	}

	return nil
}

// Generic operations for all entities
func (r *SeatingRepository) GetAllExams(ctx context.Context) ([]*Exam, error) {
	cursor, err := r.examsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var exams []*Exam
	if err := cursor.All(ctx, &exams); err != nil {
		return nil, err
	}
	return exams, nil
}

func (r *SeatingRepository) GetAllStudents(ctx context.Context) ([]*Student, error) {
	cursor, err := r.studentsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var students []*Student
	if err := cursor.All(ctx, &students); err != nil {
		return nil, err
	}
	return students, nil
}

func (r *SeatingRepository) GetAllSeatingPlans(ctx context.Context) ([]*SeatingPlan, error) {
	cursor, err := r.seatingPlansCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var plans []*SeatingPlan
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *SeatingRepository) GetAllRooms(ctx context.Context) ([]*Room, error) {
	cursor, err := r.roomsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var rooms []*Room
	if err := cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *SeatingRepository) GetAllStudentLists(ctx context.Context) ([]*StudentList, error) {
	cursor, err := r.studentListsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var studentLists []*StudentList
	if err := cursor.All(ctx, &studentLists); err != nil {
		return nil, err
	}
	return studentLists, nil
}

func (r *SeatingRepository) GetAllInvigilators(ctx context.Context) ([]*User, error) {
	filter := bson.M{"role": bson.M{"$in": []string{"admin", "staff"}}}
	cursor, err := r.usersCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var users []*User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *SeatingRepository) FindUserByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	var user User
	err := r.usersCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *SeatingRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.usersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Why: This repository abstracts all database operations for seating-related entities, making it easier to test and maintain the seating logic.
