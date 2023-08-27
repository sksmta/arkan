package cache

import (
	"container/list"
	"sync"
)

// LRUCache represents an LRU Cache
type LRUCache struct {
	capacity     int
	cache        map[int64]*list.Element
	evictionList *list.List
	mutex        sync.Mutex
}

// NewLRUCache creates a new LRU Cache
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity:     capacity,
		cache:        make(map[int64]*list.Element),
		evictionList: list.New(),
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key int64) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, ok := c.cache[key]; ok {
		// Move the element to the front of the eviction list
		c.evictionList.MoveToFront(elem)
		return elem.Value.([]byte), true
	}
	return nil, false
}

// Add adds a value to the cache
func (c *LRUCache) Add(key int64, value []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, ok := c.cache[key]; ok {
		// Update the existing element's value and move it to the front
		elem.Value = value
		c.evictionList.MoveToFront(elem)
	} else {
		// Add a new element to the cache and the front of the eviction list
		elem := c.evictionList.PushFront(value)
		c.cache[key] = elem

		// Check and remove the least recently used element if the capacity is exceeded
		if c.evictionList.Len() > c.capacity {
			c.removeOldest()
		}
	}
}

// removeOldest removes the least recently used element from the cache
func (c *LRUCache) removeOldest() {
	elem := c.evictionList.Back()
	if elem != nil {
		c.evictionList.Remove(elem)
		delete(c.cache, elem.Value.(int64))
	}
}
