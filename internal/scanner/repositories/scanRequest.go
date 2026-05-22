package repositories

import (
	"context"
	"fmt"

	"github.com/mrtdeh/scanners-management/internal/scanner/domains"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ctx = context.Background()
)

type ScanRequestRepository interface {
	Create(req domains.ScanRequest) error
	Update(req domains.ScanRequest) error
	UpdateFile(scanId string, fileReq domains.ScanRequestFile) error
	Delete(scanID string) error
	GetByID(scanID string) (*domains.ScanRequest, error)
	List(bson.M) ([]domains.ScanRequest, error)
}

type scanRequestRepo struct {
	collection *mongo.Collection
}

func NewScanRequestRepository(db *mongo.Database) ScanRequestRepository {
	return &scanRequestRepo{
		collection: db.Collection("scan_requests"),
	}
}

func (sr *scanRequestRepo) Create(req domains.ScanRequest) error {
	_, err := sr.collection.InsertOne(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create scan request: %v", err)
	}
	return nil
}

func (sr *scanRequestRepo) Update(req domains.ScanRequest) error {
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

func (sr *scanRequestRepo) UpdateFile(scanId string, fileReq domains.ScanRequestFile) error {
	ctx := context.Background()

	filter := bson.M{"scan_id": scanId}

	update := bson.M{
		"$set": bson.M{
			"files.$[elem].name": fileReq.Name,
			"files.$[elem].hash": fileReq.Hash,
			"files.$[elem].size": fileReq.Size,
			// "files.$[elem].received": fileReq.Received,
		},
	}

	arrayFilters := options.UpdateOne().SetArrayFilters([]any{
		bson.M{"elem.id": fileReq.ID},
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

func (sr *scanRequestRepo) GetByID(scanID string) (*domains.ScanRequest, error) {
	filter := bson.M{"scan_id": scanID}

	res := sr.collection.FindOne(ctx, filter)
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to get scan request: %v", res.Err())
	}
	var req domains.ScanRequest
	if err := res.Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode scan request: %v", err)
	}
	return &req, nil
}

func (sr *scanRequestRepo) List(filter bson.M) ([]domains.ScanRequest, error) {
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

	var requests []domains.ScanRequest
	if err = cursor.All(ctx, &requests); err != nil {
		return nil, fmt.Errorf("failed to decode scan requests: %v", err)
	}

	return requests, nil
}
