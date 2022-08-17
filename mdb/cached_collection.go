package mdb

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/madkins23/go-utils/check"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CachedCollection caches Mongo-stored objects so that the same object is always returned.
// This is most useful for objects that change rarely.
//
// TODO(mAdkins): Should the change stream be used to clear cache items?
type CachedCollection[C Cacheable] struct {
	TypedCollection[C]
	cache       map[string]C
	expireAfter time.Duration
	lock        sync.Mutex
}

// ConnectCachedCollection creates a new cached collection object with the specified collection definition.
func ConnectCachedCollection[C Cacheable](
	access *Access, definition *CollectionDefinition, expireAfter time.Duration) (*CachedCollection[C], error) {
	collection := &CachedCollection[C]{
		cache:       make(map[string]C),
		expireAfter: expireAfter,
	}
	if err := access.CollectionConnect(&collection.Collection, definition); err != nil {
		return nil, fmt.Errorf("connecting cached collection: %w", err)
	}
	return collection, nil
}

// Cacheable must be searchable and loadable.
type Cacheable interface {
	Searchable
	Expirable
}

// Expirable defines behavior
type Expirable interface {
	ExpireAfter(duration time.Duration)
	Expired() bool
}

// Searchable may be used just for searching for a cached item.
// This supports keys that are not complete items.
type Searchable interface {
	CacheKey() string
	Filter() bson.D
}

// Delete object in cache and DB.
// Because items are Cacheable (and therefore Searchable) the item itself is passed instead of a filter.
func (c *CachedCollection[Cacheable]) Delete(item Searchable, idempotent bool) error {
	c.lock.Lock()
	delete(c.cache, item.CacheKey())
	c.lock.Unlock()
	return c.Collection.Delete(item.Filter(), idempotent)
}

// DeleteAll all objects in cache and DB.
func (c *CachedCollection[Cacheable]) DeleteAll() error {
	c.lock.Lock()
	c.cache = make(map[string]Cacheable)
	c.lock.Unlock()
	return c.Collection.DeleteAll()
}

// Find a cacheable object in either cache or database.
func (c *CachedCollection[C]) Find(searchFor Searchable) (C, error) {
	var item C
	var found bool
	c.lock.Lock()
	defer c.lock.Unlock()
	cacheKey := searchFor.CacheKey()
	if item, found = c.cache[cacheKey]; found {
		remove := false
		if check.IsZero[C](item) {
			remove = true
		} else {
			if item.Expired() {
				delete(c.cache, cacheKey)
				found = false
			}
		}
		if remove {
			delete(c.cache, cacheKey)
			found = false
		}
	}

	if !found {
		result := c.FindOne(c.ctx, searchFor.Filter())
		if err := result.Err(); err != nil {
			if IsNotFound(err) {
				return item, fmt.Errorf("no item '%v': %w", searchFor, err)
			}
			return item, fmt.Errorf("find item '%v': %w", searchFor, err)
		}
		newItem := new(C)
		if err := result.Decode(newItem); err != nil {
			return item, fmt.Errorf("decode item: %w", err)
		}
		c.cache[cacheKey] = *newItem
		return *newItem, nil
	}

	return item, nil
}

// FindOrCreate returns an existing cacheable object or creates it if it does not already exist.
func (c *CachedCollection[Cacheable]) FindOrCreate(cacheItem Cacheable) (Cacheable, error) {
	// Can't inherit from TypedCollection here, must redo the algorithm due to caching.
	item, err := c.Find(cacheItem)
	if err != nil {
		if !IsNotFound(err) {
			return item, err
		}

		err = c.Create(cacheItem)
		if err != nil {
			return item, err
		}

		item, err = c.Find(cacheItem)
		if err != nil {
			return item, fmt.Errorf("find just created item: %w", err)
		}
	}

	return item, nil
}

// InvalidateByPrefix removes items from the cache if the item key has the specified prefix.
func (c *CachedCollection[T]) InvalidateByPrefix(prefix string) {
	c.lock.Lock()
	for k := range c.cache {
		if strings.HasPrefix(k, prefix) {
			delete(c.cache, k)
		}
	}
	c.lock.Unlock()
}

// Iterate over a set of items, applying the specified function to each one.
func (c *CachedCollection[T]) Iterate(filter bson.D, fn func(item T) error) error {
	if cursor, err := c.Collection.Collection.Find(c.ctx, filter); err != nil {
		return fmt.Errorf("find items: %w", err)
	} else {
		item := new(T)
		for cursor.Next(c.ctx) {
			if err := cursor.Decode(item); err != nil {
				return fmt.Errorf("decode item: %w", err)
			}

			if err := fn(*item); err != nil {
				return fmt.Errorf("apply function: %w", err)
			}
		}
	}

	return nil
}

// Replace entire item referenced by filter with specified item.
// Unlike Collection and TypedCollection Replace methods
// the filter here must be a typed item in order to properly clear the cache entry.
func (c *CachedCollection[T]) Replace(filter, item T, opts ...*options.UpdateOptions) error {
	return c.Update(filter, bson.M{"$set": item}, opts...)
}

// Update item referenced by filter by applying update operator expressions.
// If the filter matches more than one document mongo-go-driver will choose one to update.
func (c *CachedCollection[T]) Update(filter T, changes interface{}, opts ...*options.UpdateOptions) error {
	err := c.Collection.Update(filter.Filter(), changes, opts...)
	if err != nil {
		return fmt.Errorf("basic replace: %w", err)
	}

	filterKey := filter.CacheKey()
	c.lock.Lock()
	delete(c.cache, filterKey)
	c.lock.Unlock()

	// If the changes affect the cache key there is no convenient way to figure
	// out what the new cache key would be in order to delete it from the cache.

	return nil
}
