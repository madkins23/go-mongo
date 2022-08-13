package mdb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection struct {
	*Access
	*mongo.Collection
	ctx context.Context
}

// ConnectCollection creates a new collection object with the specified collection definition.
func ConnectCollection(access *Access, definition *CollectionDefinition) (*Collection, error) {
	collection := &Collection{}
	if err := access.CollectionConnect(collection, definition); err != nil {
		return nil, fmt.Errorf("connecting collection: %w", err)
	}
	return collection, nil
}

func (c *Collection) ContextWithTimeout() (context.Context, context.CancelFunc) {
	return c.Access.ContextWithTimeout(c.Access.config.Collection)
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
	_, err := c.DeleteMany(c.ctx, NoFilter())
	if err != nil {
		return fmt.Errorf("delete all: %w", err)
	}
	return nil
}

// Drop collection.
func (c *Collection) Drop() error {
	ctx, cancelFn := c.ContextWithTimeout()
	defer cancelFn()
	return c.Collection.Drop(ctx)
}

// Find an item in the database and return it as a blank interface.
// The result will likely contain bson objects.
func (c *Collection) Find(filter bson.D) (interface{}, error) {
	var item interface{}
	if err := c.FindOne(c.ctx, filter).Decode(&item); err != nil {
		if IsNotFound(err) {
			return nil, fmt.Errorf("no item '%v': %w", filter, err)
		}
		return nil, fmt.Errorf("find item '%v': %w", filter, err)
	}

	return item, nil
}

// FindOrCreate returns an existing object or creates it if it does not already exist.
// The filter must correctly find the object as a second Find is done after any necessary creation.
func (c *Collection) FindOrCreate(filter bson.D, item interface{}) (interface{}, error) {
	found, err := c.Find(filter)
	if err != nil {
		if !IsNotFound(err) {
			return found, err
		}

		err = c.Create(item)
		if err != nil {
			return found, err
		}

		found, err = c.Find(filter)
		if err != nil {
			return found, fmt.Errorf("find just created item: %w", err)
		}
	}

	return found, nil
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

// StringValuesFor returns an array of distinct string values for the specified filter and field.
func (c *Collection) StringValuesFor(field string, filter bson.D) ([]string, error) {
	if filter == nil {
		filter = NoFilter()
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

var errNoItemMatch = errors.New("no matching item")
var errNoItemModified = errors.New("no modified item")

// Replace entire item referenced by filter with specified item.
// If the filter matches more than one document Mongo will choose one to update.
func (c *Collection) Replace(filter, item interface{}, opts ...*options.UpdateOptions) error {
	return c.Update(filter, bson.M{"$set": item}, opts...)
}

// Update item referenced by filter by applying update operator expressions.
// If the filter matches more than one document Mongo will choose one to update.
func (c *Collection) Update(filter, operators interface{}, opts ...*options.UpdateOptions) error {
	result, err := c.UpdateOne(c.Context(), filter, operators, opts...)
	if err != nil {
		return fmt.Errorf("replace item: %w", err)
	} else if result.MatchedCount < 1 && result.UpsertedCount < 1 {
		return errNoItemMatch
	} else if result.ModifiedCount < 1 && result.UpsertedCount < 1 {
		// Not sure how to test this,  may never happen.
		return errNoItemModified
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////

// NoFilter returns an empty bson.D object for use as an empty filter.
func NoFilter() bson.D {
	return bson.D{}
}
