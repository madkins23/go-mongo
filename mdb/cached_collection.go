package mdb

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// CachedCollection Mongo-stored objects so that the same object is always returned.
// This is most useful for objects that change rarely.
type CachedCollection struct {
	*Access
	*mongo.Collection
	cache       map[string]Cacheable
	ctx         context.Context
	itemType    reflect.Type // if only we had generics
	expireAfter time.Duration
}

func NewCachedCollection(
	access *Access, collection *mongo.Collection,
	ctx context.Context, example Cacheable, expireAfter time.Duration) *CachedCollection {
	exampleType := reflect.TypeOf(example)
	if exampleType != nil && exampleType.Kind() == reflect.Ptr {
		exampleType = exampleType.Elem()
	}
	return &CachedCollection{
		Access:      access,
		Collection:  collection,
		cache:       make(map[string]Cacheable),
		ctx:         ctx,
		itemType:    exampleType,
		expireAfter: expireAfter,
	}
}

// Cacheable must be searchable, storable and able to be recreated and expired.
type Cacheable interface {
	Searchable
	ExpireAfter(duration time.Duration)
	Expired() bool
	Realize() error
}

// Searchable may be used just for searching for a cached item.
// This supports keys that are not complete items.
type Searchable interface {
	CacheKey() (string, error)
	Filter() bson.D
}

// Create object in DB but not cache.
func (c *CachedCollection) Create(item interface{}) error {
	if _, err := c.InsertOne(c.ctx, item); err != nil {
		return fmt.Errorf("insert item: %w", err)
	}

	return nil
}

// Delete object in cache and DB.
func (c *CachedCollection) Delete(item Searchable, idempotent bool) error {
	if cacheKey, err := item.CacheKey(); err != nil {
		return fmt.Errorf("cache key: %w", err)
	} else {
		delete(c.cache, cacheKey)
	}

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

	cacheKey, err := searchFor.CacheKey()
	if err != nil {
		return nil, fmt.Errorf("cache key: %w", err)
	}
	if item, found = c.cache[cacheKey]; found {
		if item == nil || item.Expired() {
			delete(c.cache, cacheKey)
			found = false
		}
	}

	if !found {
		item = c.instantiate()
		err := c.FindOne(c.ctx, searchFor.Filter()).Decode(item)
		if err != nil {
			if c.NotFound(err) {
				return nil, fmt.Errorf("no item '%s': %w", cacheKey, err)
			}
			return nil, fmt.Errorf("find item '%s': %w", cacheKey, err)
		}

		if err = item.Realize(); err != nil {
			return nil, fmt.Errorf("init from: %w", err)
		}
		item.ExpireAfter(c.expireAfter)
		c.cache[cacheKey] = item
	}

	return item, nil
}

// FindOrCreate returns an existing cacheable object or creates it if it does not already exist.
func (c *CachedCollection) FindOrCreate(cacheItem Cacheable) (Cacheable, error) {
	item, err := c.Find(cacheItem)
	if err != nil {
		if !c.NotFound(err) {
			return nil, err
		}

		err = c.Create(cacheItem)
		if err != nil {
			return nil, err
		}

		item, err = c.Find(cacheItem)
		if err != nil {
			return nil, fmt.Errorf("find just created item: %w", err)
		}
	}

	return item, nil
}

func (c *CachedCollection) instantiate() Cacheable {
	// TODO: can we assume that the item type will return an Cacheable?
	return reflect.New(c.itemType).Interface().(Cacheable)
}
