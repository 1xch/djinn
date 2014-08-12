package jingo

import (
	"container/list"
	"html/template"
	"sync"
	"time"
)

type (
	entry struct {
		key           string
		t             *template.Template
		time_accessed time.Time
	}

	TLRUCache struct {
		sync.RWMutex
		MaxEntries int
		list       *list.List
		cache      map[string]*list.Element
	}
)

func NewTLRUCache(maxentries int) *TLRUCache {
	return &TLRUCache{
		list:       list.New(),
		cache:      make(map[string]*list.Element),
		MaxEntries: maxentries,
	}
}

func (c *TLRUCache) Add(key string, tmpl *template.Template) {
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
		c.RemoveOldest()
	}
}

func (c *TLRUCache) Get(key string) (tmpl *template.Template, ok bool) {
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

func (c *TLRUCache) Remove(key string) {
	c.Lock()
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
	c.Unlock()
}

func (c *TLRUCache) Clear() {
	c.Lock()
	c.list.Init()
	c.cache = make(map[string]*list.Element)
	c.MaxEntries = 0
	c.Unlock()
}

func (c *TLRUCache) RemoveOldest() {
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

func (c *TLRUCache) removeElement(e *list.Element) {
	c.list.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
}

func (c *TLRUCache) moveToFront(element *list.Element) {
	c.list.MoveToFront(element)
	element.Value.(*entry).time_accessed = time.Now()
}

func (c *TLRUCache) addNew(key string, tmpl *template.Template) {
	c.Lock()
	newEntry := &entry{key, tmpl, time.Now()}
	element := c.list.PushFront(newEntry)
	c.cache[key] = element
	c.Unlock()
}
