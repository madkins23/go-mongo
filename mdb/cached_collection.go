package mdb

import (
	"fmt"
	"strings"
	"time"

	"github.com/madkins23/go-utils/check"
	"go.mongodb.org/mongo-driver/bson"
)

// CachedCollection caches Mongo-stored objects so that the same object is always returned.
// This is most useful for objects that change rarely.
//
// TODO: Should this be made thread-safe?
type CachedCollection[C Cacheable] struct {
	Collection
	cache       map[string]C
	expireAfter time.Duration
}

func NewCachedCollection[C Cacheable](collection *Collection, expireAfter time.Duration) *CachedCollection[C] {
	return &CachedCollection[C]{
		Collection:  *collection,
		cache:       make(map[string]C),
		expireAfter: expireAfter,
	}
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

// Create item in DB and cache.
func (c *CachedCollection[Cacheable]) Create(item Cacheable) error {
	if _, err := c.InsertOne(c.ctx, item); err != nil {
		return fmt.Errorf("insert item: %w", err)
	}
	item.ExpireAfter(c.expireAfter)
	c.cache[item.CacheKey()] = item
	return nil
}

// Delete object in cache and DB.
func (c *CachedCollection[Cacheable]) Delete(item Searchable, idempotent bool) error {
	delete(c.cache, item.CacheKey())
	return c.Collection.Delete(item.Filter(), idempotent)
}

// DeleteAll all objects in cache and DB.
func (c *CachedCollection[Cacheable]) DeleteAll() error {
	// TODO: Why does this compile when c.cache is defined as map[string]C in the struct?
	c.cache = make(map[string]Cacheable)
	return c.Collection.DeleteAll()
}

// Find a cacheable object in either cache or database.
func (c *CachedCollection[C]) Find(searchFor Searchable) (C, error) {
	var item C
	var found bool
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
		newItem := new(C)
		err := c.FindOne(c.ctx, searchFor).Decode(newItem)
		if err != nil {
			if c.IsNotFound(err) {
				return item, fmt.Errorf("no item '%v': %w", searchFor, err)
			}
			return item, fmt.Errorf("find item '%v': %w", searchFor, err)
		}

		return *newItem, nil
	}

	return item, nil
}

// FindOrCreate returns an existing cacheable object or creates it if it does not already exist.
func (c *CachedCollection[Cacheable]) FindOrCreate(cacheItem Cacheable) (Cacheable, error) {
	item, err := c.Find(cacheItem)
	if err != nil {
		if !c.IsNotFound(err) {
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
	for k := range c.cache {
		if strings.HasPrefix(k, prefix) {
			delete(c.cache, k)
		}
	}
}
