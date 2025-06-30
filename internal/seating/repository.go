package seating

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SeatingRepository handles DB operations for seating-related entities.
type SeatingRepository struct {
	studentsCollection     *mongo.Collection
	roomsCollection        *mongo.Collection
	examsCollection        *mongo.Collection
	invigilatorsCollection *mongo.Collection
	seatingPlansCollection *mongo.Collection
}

// NewSeatingRepository creates a new repository for seating operations.
func NewSeatingRepository(db *mongo.Database) *SeatingRepository {
	return &SeatingRepository{
		studentsCollection:     db.Collection("students"),
		roomsCollection:        db.Collection("rooms"),
		examsCollection:        db.Collection("exams"),
		invigilatorsCollection: db.Collection("invigilators"),
		seatingPlansCollection: db.Collection("seating_plans"),
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

// Why: This repository abstracts all database operations for seating-related entities, making it easier to test and maintain the seating logic.
