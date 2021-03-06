package mdb

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

// Realizable can be initialized after loading.
type Realizable interface {
	Realize() error
}

// TypedCollection uses reflection to properly create objects returned from Mongo.
type TypedCollection struct {
	*Collection
	itemType reflect.Type // if only we had generics
}

func NewTypedCollection(collection *Collection, example interface{}) *TypedCollection {
	exampleType := reflect.TypeOf(example)
	if exampleType != nil && exampleType.Kind() == reflect.Ptr {
		exampleType = exampleType.Elem()
	}
	return &TypedCollection{
		Collection: collection,
		itemType:   exampleType,
	}
}

// Find an item in the database.
// Will return an interface to an item of the collection's type.
// If the item is Realizable then it will be realized before returning.
func (c *TypedCollection) Find(filter bson.D) (interface{}, error) {
	item := c.Instantiate()
	err := c.FindOne(c.ctx, filter).Decode(item)
	if err != nil {
		if c.NotFound(err) {
			return nil, fmt.Errorf("no item '%v': %w", filter, err)
		}
		return nil, fmt.Errorf("find item '%v': %w", filter, err)
	}

	if realizable, ok := item.(Realizable); ok {
		if err := realizable.Realize(); err != nil {
			return nil, fmt.Errorf("realize item: %w", err)
		}
	}

	return item, nil
}

// Iterate over a set of items, applying the specified function to each one.
// The items passed to the function will likely contain bson objects.
func (c *TypedCollection) Iterate(filter bson.D, fn func(item interface{}) error) error {
	if cursor, err := c.Collection.Collection.Find(c.ctx, filter); err != nil {
		return fmt.Errorf("find items: %w", err)
	} else {
		item := c.Instantiate()
		for cursor.Next(c.ctx) {
			if err := cursor.Decode(item); err != nil {
				return fmt.Errorf("decode item: %w", err)
			}

			if realizable, ok := item.(Realizable); ok {
				if err := realizable.Realize(); err != nil {
					return fmt.Errorf("realize item: %w", err)
				}
			}

			if err := fn(item); err != nil {
				return fmt.Errorf("apply function: %w", err)
			}
		}
	}

	return nil
}

// Instantiate the item specified by the item type.
func (c *TypedCollection) Instantiate() interface{} {
	return reflect.New(c.itemType).Interface()
}
