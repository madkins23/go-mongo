package mdb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	errMissingCollectionName = errors.New("no collection name argument")
)

// CollectionExists checks to see if a specific collection already exists.
func (a *Access) CollectionExists(name string) (bool, error) {
	if name == "" {
		return false, errMissingCollectionName
	}

	names, err := a.database.ListCollectionNames(context.TODO(), bson.M{"name": name})
	if err != nil {
		return false, fmt.Errorf("getting collection names: %w", err)
	}

	exists := false
	for _, collName := range names {
		if collName == name {
			exists = true
			break
		}
	}

	return exists, nil
}

// CollectionFinisher provides a way to add special processing when creating a collection.
type CollectionFinisher func(access *Access, collection *mongo.Collection) error

// Collection acquires the named collection, creating it if necessary.
func (a *Access) Collection(collectionName string, validatorJSON string, finishers ...CollectionFinisher) (*mongo.Collection, error) {
	if exists, err := a.CollectionExists(collectionName); err != nil {
		return nil, fmt.Errorf("does collection '%s' exist: %w", collectionName, err)
	} else if exists {
		// Collection already exists, just return it.
		return a.database.Collection(collectionName), nil
	}

	// Add option for validator JSON if it is provided.
	opts := make([]*options.CreateCollectionOptions, 0)
	if validatorJSON != "" {
		var validator interface{}
		if err := bson.UnmarshalExtJSON([]byte(validatorJSON), false, &validator); err != nil {
			return nil, fmt.Errorf("unmarshal validator for collection: %w", err)
		}
		opts = append(opts, &options.CreateCollectionOptions{Validator: validator})
	}

	// Create collection.
	err := a.database.CreateCollection(context.TODO(), collectionName, opts...)
	if err != nil {
		if cmdErr, ok := err.(mongo.CommandError); !ok || !cmdErr.HasErrorLabel("NamespaceExists") {
			return nil, fmt.Errorf("create collection: %w", err)
		}
	}
	collection := a.database.Collection(collectionName)

	// Run finishers on the collection.
	for i, finisher := range finishers {
		if err = finisher(a, collection); err != nil {
			return nil, fmt.Errorf("collection finisher #%d: %w", i, err)
		}
	}

	return collection, nil
}
