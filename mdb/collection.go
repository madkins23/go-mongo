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

	ctx, cancel := a.ContextWithTimeout(a.config.Timeout.Collection)
	defer cancel()
	names, err := a.database.ListCollectionNames(ctx, bson.M{"name": name})
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
type CollectionFinisher func(access *Access, collection *Collection) error

type Collection struct {
	*Access
	*mongo.Collection
	ctx context.Context
}

// Collection acquires the named collection, creating it if necessary.
func (a *Access) Collection(
	ctx context.Context, collectionName string, validatorJSON string, finishers ...CollectionFinisher) (*Collection, error) {
	if exists, err := a.CollectionExists(collectionName); err != nil {
		return nil, fmt.Errorf("does collection '%s' exist: %w", collectionName, err)
	} else if exists {
		// Collection already exists, just return it.
		return &Collection{Access: a, Collection: a.database.Collection(collectionName)}, nil
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
	ctx, cancel := a.ContextWithTimeout(a.config.Timeout.Collection)
	defer cancel()
	err := a.database.CreateCollection(ctx, collectionName, opts...)
	if err != nil {
		if cmdErr, ok := err.(mongo.CommandError); !ok || !cmdErr.HasErrorLabel("NamespaceExists") {
			return nil, fmt.Errorf("create collection: %w", err)
		}
	}
	collection := &Collection{
		Access:     a,
		Collection: a.database.Collection(collectionName),
	}
	a.Info("Created collection " + collection.Name())

	// Run finishers on the collection.
	for i, finisher := range finishers {
		if err = finisher(a, collection); err != nil {
			return nil, fmt.Errorf("collection finisher #%d: %w", i, err)
		}
	}

	return collection, nil
}

// Count documents in collection matching filter.
func (c *Collection) Count(filter bson.D) (int64, error) {
	if count, err := c.Collection.CountDocuments(context.TODO(), filter); err != nil {
		return 0, fmt.Errorf("insert item: %w", err)
	} else {
		return count, nil
	}
}

// Create item in DB.
func (c *Collection) Create(item interface{}) error {
	if _, err := c.InsertOne(c.ctx, item); err != nil {
		return fmt.Errorf("insert item: %w", err)
	}

	return nil
}

// Delete item from DB.
// Set idempotent to true to avoid errors if the item does not exist.
func (c *Collection) Delete(filter bson.D, idempotent bool) error {
	result, err := c.DeleteOne(c.ctx, filter)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if result.DeletedCount > 1 || (result.DeletedCount == 0 && !idempotent) {
		// Should have deleted a single item or none if idempotent flag set.
		return fmt.Errorf("deleted %d items", result.DeletedCount)
	}

	return nil
}

// DeleteAll items from this collection.
func (c *Collection) DeleteAll() error {
	_, err := c.DeleteMany(c.ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("delete all: %w", err)
	}
	return nil
}

// Find an item in the database and return it as a blank interface.
// The result will likely contain bson objects.
func (c *Collection) Find(filter bson.D) (interface{}, error) {
	var item interface{}
	if err := c.FindOne(c.ctx, filter).Decode(&item); err != nil {
		if c.NotFound(err) {
			return nil, fmt.Errorf("no item '%v': %w", filter, err)
		}
		return nil, fmt.Errorf("find item '%v': %w", filter, err)
	}

	return item, nil
}

// Iterate over a set of items, applying the specified function to each one.
// The items passed to the function will likely contain bson objects.
func (c *Collection) Iterate(filter bson.D, fn func(item interface{}) error) error {
	if cursor, err := c.Collection.Find(c.ctx, filter); err != nil {
		return fmt.Errorf("find items: %w", err)
	} else {
		var item interface{}
		for cursor.Next(c.ctx) {
			if err := cursor.Decode(&item); err != nil {
				return fmt.Errorf("decode item: %w", err)
			} else if err := fn(item); err != nil {
				return fmt.Errorf("apply function: %w", err)
			}
		}
	}

	return nil
}

var errNotString = errors.New("value not a string")

func (c *Collection) StringValuesFor(field string, filter bson.D) ([]string, error) {
	if filter == nil {
		filter = bson.D{}
	}
	values, err := c.Distinct(c.Context(), field, filter)
	if err != nil {
		return nil, fmt.Errorf("distinct values: %w", err)
	}

	var ok bool
	length := len(values)
	result := make([]string, length)
	for i := 0; i < length; i++ {
		if result[i], ok = values[i].(string); !ok {
			return nil, errNotString
		}
	}

	return result, nil
}
