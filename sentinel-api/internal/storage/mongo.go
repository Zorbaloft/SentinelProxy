package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	db *mongo.Database
}

func New(mongoURI string) (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("mongo connection failed: %w", err)
	}

	db := client.Database("sentinel")
	return &Storage{db: db}, nil
}

func (s *Storage) GetLogs(ctx context.Context, filter bson.M, limit int64, cursor string) ([]bson.M, string, error) {
	collection := s.db.Collection("logs")

	// Add cursor filter if provided
	if cursor != "" {
		objID, err := primitive.ObjectIDFromHex(cursor)
		if err == nil {
			filter["_id"] = bson.M{"$gt": objID}
		}
	}

	opts := options.Find().SetLimit(limit).SetSort(bson.M{"timestamp": -1})
	cursorResult, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	defer cursorResult.Close(ctx)

	var logs []bson.M
	if err := cursorResult.All(ctx, &logs); err != nil {
		return nil, "", err
	}

	nextCursor := ""
	if len(logs) > 0 {
		lastID := logs[len(logs)-1]["_id"]
		if id, ok := lastID.(primitive.ObjectID); ok {
			nextCursor = id.Hex()
		}
	}

	return logs, nextCursor, nil
}

func (s *Storage) GetIncidents(ctx context.Context, status string) ([]bson.M, error) {
	collection := s.db.Collection("incidents")
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}

	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.M{"timestamp": -1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var incidents []bson.M
	if err := cursor.All(ctx, &incidents); err != nil {
		return nil, err
	}

	return incidents, nil
}

func (s *Storage) CloseIncident(ctx context.Context, id string) error {
	collection := s.db.Collection("incidents")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"status": "closed"}})
	return err
}

func (s *Storage) GetRules(ctx context.Context) ([]bson.M, error) {
	collection := s.db.Collection("rules")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []bson.M
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

func (s *Storage) CreateRule(ctx context.Context, rule bson.M) (primitive.ObjectID, error) {
	collection := s.db.Collection("rules")
	result, err := collection.InsertOne(ctx, rule)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (s *Storage) UpdateRule(ctx context.Context, id string, rule bson.M) error {
	collection := s.db.Collection("rules")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": rule})
	return err
}

func (s *Storage) DeleteRule(ctx context.Context, id string) error {
	collection := s.db.Collection("rules")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
