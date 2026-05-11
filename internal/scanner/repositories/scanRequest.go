package repositories

import (
	"context"
	"fmt"

	"github.com/mrtdeh/scanners-management/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ctx = context.Background()
)

type ScanRequestRepository interface {
	Create(req model.ScanRequest) error
	Update(req model.ScanRequest) error
	UpdateFile(scanId string, fileReq model.ScanRequestFile) error
	Delete(scanID string) error
	GetByID(scanID string) (*model.ScanRequest, error)
	List(bson.M) ([]model.ScanRequest, error)
}

type scanRequestRepo struct {
	collection *mongo.Collection
}

func NewScanRequestRepository(db *mongo.Database) ScanRequestRepository {
	return &scanRequestRepo{
		collection: db.Collection("scan_requests"),
	}
}

func (sr *scanRequestRepo) Create(req model.ScanRequest) error {
	_, err := sr.collection.InsertOne(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create scan request: %v", err)
	}
	return nil
}

func (sr *scanRequestRepo) Update(req model.ScanRequest) error {
	filter := bson.M{"scan_id": req.ScanID}

	result, err := sr.collection.UpdateMany(ctx, filter, bson.M{"$set": req})
	if err != nil {
		return fmt.Errorf("failed to update document: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("scan request with id %s not found", req.ScanID)
	}
	return nil
}

func (sr *scanRequestRepo) UpdateFile(scanId string, fileReq model.ScanRequestFile) error {
	ctx := context.Background()

	filter := bson.M{"scan_id": scanId}

	// آپدیت فایل در آرایه files (اگر فایل با همین name و hash وجود داشت، replace کن)
	update := bson.M{
		"$set": bson.M{
			"files.$[elem].name":     fileReq.Name,
			"files.$[elem].hash":     fileReq.Hash,
			"files.$[elem].size":     fileReq.Size,
			"files.$[elem].received": fileReq.Received,
		},
	}

	arrayFilters := options.UpdateOne().SetArrayFilters([]any{
		bson.M{"elem.request_id": fileReq.RequestID},
	})

	result, err := sr.collection.UpdateOne(ctx, filter, update, arrayFilters)
	if err != nil {
		return fmt.Errorf("failed to update file: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("scan request with id %s not found or file not found", scanId)
	}

	return nil
}

func (sr *scanRequestRepo) Delete(scanID string) error {
	filter := bson.M{"scan_id": scanID}

	_, err := sr.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete scan request: %v", err)
	}
	return nil
}

func (sr *scanRequestRepo) GetByID(scanID string) (*model.ScanRequest, error) {
	filter := bson.M{"scan_id": scanID}

	res := sr.collection.FindOne(ctx, filter)
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to get scan request: %v", res.Err())
	}
	var req model.ScanRequest
	if err := res.Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode scan request: %v", err)
	}
	return &req, nil
}

func (sr *scanRequestRepo) List(filter bson.M) ([]model.ScanRequest, error) {
	if filter == nil {
		filter = bson.M{}
	}

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "started_at", Value: -1}})

	cursor, err := sr.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list scan requests: %v", err)
	}
	defer cursor.Close(ctx)

	var requests []model.ScanRequest
	if err = cursor.All(ctx, &requests); err != nil {
		return nil, fmt.Errorf("failed to decode scan requests: %v", err)
	}

	return requests, nil
}
