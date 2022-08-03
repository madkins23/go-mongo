package mdb

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

// TypedCollection uses reflection to properly create objects returned from Mongo.
type TypedCollection[T any] struct {
	Collection
}

func NewTypedCollection[T any](collection *Collection) *TypedCollection[T] {
	return &TypedCollection[T]{
		Collection: *collection,
	}
}

// Find an item in the database.
// Will return an interface to an item of the collection's type.
// If the item is Realizable then it will be realized before returning.
func (c *TypedCollection[T]) Find(filter bson.D) (*T, error) {
	item := new(T)
	err := c.FindOne(c.ctx, filter).Decode(item)
	if err != nil {
		if c.IsNotFound(err) {
			return nil, fmt.Errorf("no item '%v': %w", filter, err)
		}
		return nil, fmt.Errorf("find item '%v': %w", filter, err)
	}

	return item, nil
}

// Iterate over a set of items, applying the specified function to each one.
// The items passed to the function will likely contain bson objects.
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
