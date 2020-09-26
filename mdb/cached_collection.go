package mdb

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	errNotCacheable = errors.New("item not cacheable")
	errNotInterface = errors.New("item not an interface")
)

// CachedCollection Mongo-stored objects so that the same object is always returned.
// This is most useful for objects that change rarely.
type CachedCollection struct {
	*Access
	*mongo.Collection
	cache       map[string]Cacheable
	ctx         context.Context
	itemType    reflect.Type
	expireAfter time.Duration
}

func NewCache(
	access *Access, collection *mongo.Collection,
	ctx context.Context, example interface{}, expireAfter time.Duration) *CachedCollection {
	examType := reflect.TypeOf(example)
	if examType != nil && examType.Kind() == reflect.Ptr {
		examType = examType.Elem()
	}
	return &CachedCollection{
		Access:      access,
		Collection:  collection,
		cache:       make(map[string]Cacheable),
		ctx:         ctx,
		itemType:    examType,
		expireAfter: expireAfter,
	}
}

// Searchable may be used just for searching for a cached item.
// This supports keys that are not complete items.
type Searchable interface {
	CacheKey() string
	Filter() bson.D
}

// Cacheable must be searchable and able to be recreated and expired.
type Cacheable interface {
	Searchable
	ExpireAfter(time.Duration)
	Expired() bool
	InitFrom(stub bson.M) error
}

// Storable must be able to generate a Mongo document.
// This supports "stub" objects used just to add items.
type Storable interface {
	Document() bson.M
}

// Create object in DB but not cache.
func (c *CachedCollection) Create(item Storable) error {
	if _, err := c.InsertOne(c.ctx, item.Document()); err != nil {
		return fmt.Errorf("insert item: %w", err)
	}

	return nil
}

// Delete object in cache and DB.
func (c *CachedCollection) Delete(item Cacheable, idempotent bool) error {
	delete(c.cache, item.CacheKey())

	result, err := c.DeleteOne(c.ctx, item.Filter())
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if result.DeletedCount > 1 || (result.DeletedCount == 0 && !idempotent) {
		// Should have deleted a single item or none if idempotent flag set.
		return fmt.Errorf("deleted %d items", result.DeletedCount)
	}

	return nil
}

// Find a cacheable object in either cache or database.
func (c *CachedCollection) Find(searchFor Searchable) (Cacheable, error) {
	var found bool
	var item Cacheable

	cacheKey := searchFor.CacheKey()
	if item, found = c.cache[cacheKey]; found {
		if item == nil || item.Expired() {
			delete(c.cache, cacheKey)
			found = false
		}
	}

	if !found {
		var stub bson.M
		err := c.FindOne(c.ctx, searchFor.Filter()).Decode(&stub)
		if err != nil {
			if c.NotFound(err) {
				return nil, fmt.Errorf("no item '%s': %w", cacheKey, err)
			}
			return nil, fmt.Errorf("find item '%s': %w", cacheKey, err)
		}

		value := reflect.New(c.itemType)
		if !value.CanInterface() {
			// Would panic in next step.
			return nil, errNotInterface
		}
		var ok bool
		item, ok = value.Interface().(Cacheable)
		if !ok {
			return nil, errNotCacheable
		}
		if err = item.InitFrom(stub); err != nil {
			return nil, fmt.Errorf("init from: %w", err)
		}
		item.ExpireAfter(c.expireAfter)
		c.cache[cacheKey] = item
	}

	return item, nil
}
