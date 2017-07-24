package djinn

import (
	"container/list"
	"html/template"
	"sync"
	"time"
)

// Cache is an interface to template caching.
type Cache interface {
	Add(string, *template.Template)
	Get(string) (*template.Template, bool)
	Remove(string)
	Clear()
	On() bool
	SetCaching(bool)
}

type tlruCache struct {
	sync.RWMutex
	MaxEntries int
	list       *list.List
	cache      map[string]*list.Element
	on         bool
}

// Returns a TLRU cache interface
func TLRUCache(maxentries int, on bool) *tlruCache {
	return &tlruCache{
		MaxEntries: maxentries,
		list:       list.New(),
		cache:      make(map[string]*list.Element),
		on:         on,
	}
}

type entry struct {
	key           string
	t             *template.Template
	time_accessed time.Time
}

// Add will add the provided template with the key to the cache.
func (c *tlruCache) Add(key string, tmpl *template.Template) {
	if c.cache == nil {
		c.cache = make(map[string]*list.Element)
		c.list = list.New()
	}
	c.RLock()
	if ee, ok := c.cache[key]; ok {
		c.list.MoveToFront(ee)
		ee.Value.(*entry).t = tmpl
		return
	}
	c.RUnlock()
	c.addNew(key, tmpl)
	if c.MaxEntries != 0 && c.list.Len() > c.MaxEntries {
		c.removeOldest()
	}
}

// Get attempts to return a template corresponding to the provided key.
func (c *tlruCache) Get(key string) (tmpl *template.Template, ok bool) {
	c.RLock()
	defer c.RUnlock()
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.moveToFront(ele)
		return ele.Value.(*entry).t, true
	}
	return nil, false
}

// Remove will remove the template from the cache, indicated by the provided key
func (c *tlruCache) Remove(key string) {
	c.Lock()
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
	c.Unlock()
}

// Clear clears all entries from the cache.
func (c *tlruCache) Clear() {
	c.Lock()
	c.list.Init()
	c.cache = make(map[string]*list.Element)
	c.MaxEntries = 0
	c.Unlock()
}

func (c *tlruCache) On() bool {
	return c.on
}

func (c *tlruCache) SetCaching(to bool) {
	c.on = to
}

func (c *tlruCache) removeOldest() {
	c.Lock()
	if c.cache == nil {
		return
	}
	ele := c.list.Back()
	if ele != nil {
		c.removeElement(ele)
	}
	c.Unlock()
}

func (c *tlruCache) removeElement(e *list.Element) {
	c.list.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
}

func (c *tlruCache) moveToFront(element *list.Element) {
	c.list.MoveToFront(element)
	element.Value.(*entry).time_accessed = time.Now()
}

func (c *tlruCache) addNew(key string, tmpl *template.Template) {
	c.Lock()
	newEntry := &entry{key, tmpl, time.Now()}
	element := c.list.PushFront(newEntry)
	c.cache[key] = element
	c.Unlock()
}
