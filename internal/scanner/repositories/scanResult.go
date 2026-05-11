package repositories

import (
	"fmt"

	"github.com/mrtdeh/scanners-management/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ScanResultRepository interface {
	Create(req model.ScanResult) error
	Update(req model.ScanResult) error
	GetByFileHash(hash string) (*model.ScanResult, error)
	List(bson.M) ([]model.ScanResult, error)
}

type scanResultRepo struct {
	collection *mongo.Collection
}

func NewScanResultRepository(db *mongo.Database) ScanResultRepository {
	return &scanResultRepo{
		collection: db.Collection("scan_results"),
	}
}

func (sr *scanResultRepo) Create(req model.ScanResult) error {
	_, err := sr.collection.InsertOne(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create scan result: %v", err)
	}
	return nil
}

func (sr *scanResultRepo) Update(req model.ScanResult) error {
	filter := bson.M{"file_hash": req.FileHash}

	result, err := sr.collection.UpdateMany(ctx, filter, bson.M{"$set": req})
	if err != nil {
		return fmt.Errorf("failed to update scan result: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("scan result for file hash %s not found", req.FileHash)
	}
	return nil
}

func (sr *scanResultRepo) GetByFileHash(hash string) (*model.ScanResult, error) {
	filter := bson.M{"file_hash": hash}

	res := sr.collection.FindOne(ctx, filter)
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to get scan result: %v", res.Err())
	}
	var req model.ScanResult
	if err := res.Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode scan result: %v", err)
	}
	return &req, nil
}

func (sr *scanResultRepo) List(filter bson.M) ([]model.ScanResult, error) {
	if filter == nil {
		filter = bson.M{}
	}

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "started_at", Value: -1}})

	cursor, err := sr.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list scan results: %v", err)
	}
	defer cursor.Close(ctx)

	var results []model.ScanResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode scan results: %v", err)
	}

	return results, nil
}
