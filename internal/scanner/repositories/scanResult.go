package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/mrtdeh/scanners-management/internal/scanner/domains"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ScanResultRepository interface {
	Create(req domains.ScanResult) error
	Update(req domains.ScanResult) error
	GetByFileHash(hash string) (*domains.ScanResult, error)
	List(bson.M) ([]domains.ScanResult, error)
	PutScanResultByFileHash(hash string, result domains.ScannerResult) error
}

type scanResultRepo struct {
	collection *mongo.Collection
}

func NewScanResultRepository(db *mongo.Database) ScanResultRepository {
	return &scanResultRepo{
		collection: db.Collection("scan_results"),
	}
}
func (sr *scanResultRepo) PutScanResultByFileHash(hash string, result domains.ScannerResult) error {
	ctx := context.Background()

	c := sr.collection
	// Update engine result if existed
	filter := bson.M{
		"file_hash":      hash,
		"results.engine": result.Engine,
	}

	update := bson.M{
		"$set": bson.M{
			"results.$": result,
		},
	}

	updateResult, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// Add engine scan result if not existed
	if updateResult.MatchedCount == 0 {
		filter = bson.M{"file_hash": hash}
		update = bson.M{
			"$push": bson.M{
				"results": bson.M{
					"$each":     []domains.ScannerResult{result},
					"$position": 0,
				},
			},
			"$setOnInsert": bson.M{
				"created_at": time.Now(),
			},
		}

		opts := options.UpdateOne().SetUpsert(true)
		_, err = c.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return fmt.Errorf("push failed: %w", err)
		}
	}
	return nil
}

func (sr *scanResultRepo) Create(req domains.ScanResult) error {
	_, err := sr.collection.InsertOne(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create scan result: %v", err)
	}
	return nil
}

func (sr *scanResultRepo) Update(req domains.ScanResult) error {
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

func (sr *scanResultRepo) GetByFileHash(hash string) (*domains.ScanResult, error) {
	filter := bson.M{"file_hash": hash}

	res := sr.collection.FindOne(ctx, filter)
	if res.Err() != nil {
		return nil, fmt.Errorf("failed to get scan result: %v", res.Err())
	}
	var req domains.ScanResult
	if err := res.Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode scan result: %v", err)
	}
	return &req, nil
}

func (sr *scanResultRepo) List(filter bson.M) ([]domains.ScanResult, error) {
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

	var results []domains.ScanResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode scan results: %v", err)
	}

	return results, nil
}
