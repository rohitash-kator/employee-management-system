package mongodb

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rohitashk/golang-rest-api/internal/domain"
	domainEmployee "github.com/rohitashk/golang-rest-api/internal/domain/employee"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EmployeeRepository struct {
	coll *mongo.Collection
}

func NewEmployeeRepository(db *mongo.Database) *EmployeeRepository {
	return &EmployeeRepository{coll: db.Collection("employees")}
}

type employeeDoc struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	FirstName  string             `bson:"first_name"`
	LastName   string             `bson:"last_name"`
	Email      string             `bson:"email"`
	Department string             `bson:"department"`
	Position   string             `bson:"position"`
	Salary     float64            `bson:"salary"`
	Status     string             `bson:"status"`
	CreatedAt  time.Time          `bson:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at"`
}

func (r *EmployeeRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_email"),
		},
		{
			Keys:    bson.D{{Key: "department", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("dept_status"),
		},
	})
	if err != nil {
		return fmt.Errorf("create indexes: %w", err)
	}
	return nil
}

func (r *EmployeeRepository) Create(ctx context.Context, e *domainEmployee.Employee) error {
	doc := employeeDoc{
		FirstName:  e.FirstName,
		LastName:   e.LastName,
		Email:      strings.ToLower(strings.TrimSpace(e.Email)),
		Department: e.Department,
		Position:   e.Position,
		Salary:     e.Salary,
		Status:     string(e.Status),
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}

	res, err := r.coll.InsertOne(ctx, doc)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.Conflict("employee with this email already exists")
		}
		return domain.Internal("failed to create employee", err)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return domain.Internal("failed to parse inserted id", errors.New("unexpected inserted id type"))
	}
	e.ID = oid.Hex()
	return nil
}

func (r *EmployeeRepository) GetByID(ctx context.Context, id string) (*domainEmployee.Employee, error) {
	oid, err := parseObjectID(id)
	if err != nil {
		return nil, err
	}

	var doc employeeDoc
	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, domain.Internal("failed to fetch employee", err)
	}
	return toDomain(doc), nil
}

func (r *EmployeeRepository) GetByEmail(ctx context.Context, email string) (*domainEmployee.Employee, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, domain.Validation("email is required")
	}

	var doc employeeDoc
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, domain.Internal("failed to fetch employee", err)
	}
	return toDomain(doc), nil
}

func (r *EmployeeRepository) List(ctx context.Context, filter domainEmployee.ListFilter, page domainEmployee.ListPage) ([]domainEmployee.Employee, int64, error) {
	q := bson.M{}
	if filter.Department != nil && strings.TrimSpace(*filter.Department) != "" {
		q["department"] = strings.TrimSpace(*filter.Department)
	}
	if filter.Status != nil && *filter.Status != "" {
		q["status"] = string(*filter.Status)
	}
	if filter.Query != nil && strings.TrimSpace(*filter.Query) != "" {
		escaped := regexp.QuoteMeta(strings.TrimSpace(*filter.Query))
		re := primitive.Regex{Pattern: escaped, Options: "i"}
		q["$or"] = []bson.M{
			{"first_name": re},
			{"last_name": re},
			{"email": re},
		}
	}

	total, err := r.coll.CountDocuments(ctx, q)
	if err != nil {
		return nil, 0, domain.Internal("failed to count employees", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(page.Limit).
		SetSkip(page.Offset)

	cur, err := r.coll.Find(ctx, q, opts)
	if err != nil {
		return nil, 0, domain.Internal("failed to list employees", err)
	}
	defer cur.Close(ctx)

	out := make([]domainEmployee.Employee, 0)
	for cur.Next(ctx) {
		var doc employeeDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, 0, domain.Internal("failed to decode employee", err)
		}
		out = append(out, *toDomain(doc))
	}
	if err := cur.Err(); err != nil {
		return nil, 0, domain.Internal("failed to iterate employees", err)
	}

	return out, total, nil
}

func (r *EmployeeRepository) Update(ctx context.Context, e *domainEmployee.Employee) error {
	oid, err := parseObjectID(e.ID)
	if err != nil {
		return err
	}

	set := bson.M{
		"first_name": e.FirstName,
		"last_name":  e.LastName,
		"email":      strings.ToLower(strings.TrimSpace(e.Email)),
		"department": e.Department,
		"position":   e.Position,
		"salary":     e.Salary,
		"status":     string(e.Status),
		"updated_at": e.UpdatedAt,
	}

	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": set})
	if err != nil {
		if isDuplicateKey(err) {
			return domain.Conflict("employee with this email already exists")
		}
		return domain.Internal("failed to update employee", err)
	}
	if res.MatchedCount == 0 {
		return domain.NotFound("employee not found")
	}
	return nil
}

func (r *EmployeeRepository) Delete(ctx context.Context, id string) error {
	oid, err := parseObjectID(id)
	if err != nil {
		return err
	}

	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return domain.Internal("failed to delete employee", err)
	}
	if res.DeletedCount == 0 {
		return domain.NotFound("employee not found")
	}
	return nil
}

func toDomain(doc employeeDoc) *domainEmployee.Employee {
	return &domainEmployee.Employee{
		ID:         doc.ID.Hex(),
		FirstName:  doc.FirstName,
		LastName:   doc.LastName,
		Email:      doc.Email,
		Department: doc.Department,
		Position:   doc.Position,
		Salary:     doc.Salary,
		Status:     domainEmployee.Status(doc.Status),
		CreatedAt:  doc.CreatedAt,
		UpdatedAt:  doc.UpdatedAt,
	}
}

func parseObjectID(id string) (primitive.ObjectID, error) {
	oid, err := primitive.ObjectIDFromHex(strings.TrimSpace(id))
	if err != nil {
		return primitive.NilObjectID, domain.Validation("invalid id")
	}
	return oid, nil
}
