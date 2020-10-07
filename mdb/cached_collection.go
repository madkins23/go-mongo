package mdb

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// CachedCollection Mongo-stored objects so that the same object is always returned.
// This is most useful for objects that change rarely.
type CachedCollection struct {
	*Collection
	cache       map[string]Cacheable
	itemType    reflect.Type // if only we had generics
	expireAfter time.Duration
}

func NewCachedCollection(
	collection *Collection, example Cacheable, expireAfter time.Duration) *CachedCollection {
	exampleType := reflect.TypeOf(example)
	if exampleType != nil && exampleType.Kind() == reflect.Ptr {
		exampleType = exampleType.Elem()
	}
	return &CachedCollection{
		Collection:  collection,
		cache:       make(map[string]Cacheable),
		itemType:    exampleType,
		expireAfter: expireAfter,
	}
}

// Cacheable must be searchable and loadable.
type Cacheable interface {
	Searchable
	Loadable
}

// Loadable may be loaded and then realized from stored fields.
type Loadable interface {
	ExpireAfter(duration time.Duration)
	Expired() bool
	Realize() error
}

// Searchable may be used just for searching for a cached item.
// This supports keys that are not complete items.
type Searchable interface {
	CacheKey() string
	Filter() bson.D
}

// Delete object in cache and DB.
func (c *CachedCollection) Delete(item Searchable, idempotent bool) error {
	delete(c.cache, item.CacheKey())
	return c.Collection.Delete(item.Filter(), idempotent)
}

// Find a cacheable object in either cache or database.
func (c *CachedCollection) Find(searchFor Searchable) (Cacheable, error) {
	var found bool
	var item Cacheable

	cacheKey := searchFor.CacheKey()
	if item, found = c.cache[searchFor.CacheKey()]; found {
		if item == nil || item.Expired() {
			delete(c.cache, cacheKey)
			found = false
		}
	}

	if !found {
		item = c.Instantiate()
		err := c.FindOne(c.ctx, searchFor.Filter()).Decode(item)
		if err != nil {
			if c.NotFound(err) {
				return nil, fmt.Errorf("no item '%s': %w", cacheKey, err)
			}
			return nil, fmt.Errorf("find item '%s': %w", cacheKey, err)
		}

		if err = item.Realize(); err != nil {
			return nil, fmt.Errorf("realize item: %w", err)
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

// Instantiate the Cacheable item specified by the item type.
func (c *CachedCollection) Instantiate() Cacheable {
	// TODO: can we assume that the item type will return a Cacheable?
	return reflect.New(c.itemType).Interface().(Cacheable)
}

// InvalidateByPrefix removes items from the cache if the item key has the specified prefix.
func (c *CachedCollection) InvalidateByPrefix(prefix string) {
	for k := range c.cache {
		if strings.HasPrefix(k, prefix) {
			delete(c.cache, k)
		}
	}
}
