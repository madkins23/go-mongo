package mdb

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

// TypedCollection uses reflection to properly create objects returned from Mongo.
type TypedCollection[T any] struct {
	Collection
}

// ConnectTypedCollection creates a new typed collection object with the specified collection definition.
func ConnectTypedCollection[T any](access *Access, definition *CollectionDefinition) (*TypedCollection[T], error) {
	collection := &TypedCollection[T]{}
	if err := access.CollectionConnect(&collection.Collection, definition); err != nil {
		return nil, fmt.Errorf("connecting typed collection: %w", err)
	}
	return collection, nil
}

// Find an item in the database.
// Will return an interface to an item of the collection's type.
func (c *TypedCollection[T]) Find(filter bson.D) (*T, error) {
	result := c.FindOne(c.ctx, filter)
	if err := result.Err(); err != nil {
		if IsNotFound(err) {
			return nil, fmt.Errorf("no item '%v': %w", filter, err)
		}
		return nil, fmt.Errorf("find item '%v': %w", filter, err)
	}
	item := new(T)
	if err := result.Decode(item); err != nil {
		return item, fmt.Errorf("decode item: %w", err)
	}

	return item, nil
}

// FindOrCreate returns an existing cacheable object or creates it if it does not already exist.
func (c *TypedCollection[T]) FindOrCreate(filter bson.D, item *T) (*T, error) {
	// Can't inherit from TypedCollection here, must redo the algorithm due to typing.
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
func (c *TypedCollection[T]) Iterate(filter bson.D, fn func(item *T) error) error {
	if cursor, err := c.Collection.Collection.Find(c.ctx, filter); err != nil {
		return fmt.Errorf("find items: %w", err)
	} else {
		item := new(T)
		for cursor.Next(c.ctx) {
			if err := cursor.Decode(item); err != nil {
				return fmt.Errorf("decode item: %w", err)
			}

			if err := fn(item); err != nil {
				return fmt.Errorf("apply function: %w", err)
			}
		}
	}

	return nil
}
