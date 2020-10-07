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

// Instantiate the Cacheable item specified by the item type.
func (c *TypedCollection) Instantiate() interface{} {
	// TODO: can we assume that the item type will return a Cacheable?
	return reflect.New(c.itemType).Interface()
}
