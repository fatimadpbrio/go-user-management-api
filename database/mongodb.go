package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect() (*mongo.Client, context.Context, context.CancelFunc) {
	client, err := mongo.Connect(
		options.Client().ApplyURI("mongodb://localhost:27017"),
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	return client, ctx, cancel
}

func GetCollection(client *mongo.Client, databaseName string, collectionName string) *mongo.Collection {
	return client.Database(databaseName).Collection(collectionName)
}

func InsertOne(ctx context.Context, collection *mongo.Collection, document any) (any, error) {
	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return nil, err
	}

	return result.InsertedID, nil
}

func UpdateOne(ctx context.Context, collection *mongo.Collection, filter any, update any) error {
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func DeleteOne(ctx context.Context, collection *mongo.Collection, filter any) (int64, error) {
    result, err := collection.DeleteOne(ctx, filter)

    if err != nil {
        return 0, err
    }

    return result.DeletedCount, nil
}

func FindOne(ctx context.Context, collection *mongo.Collection, filter any, result any) error {
	return collection.FindOne(ctx, filter).Decode(result)
}
