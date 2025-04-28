//go:build !change

package lrucache

import (
	"container/list"
)

type Cache interface {
	// Get returns value associated with the key.
	//
	// The second value is a bool that is true if the key exists in the cache,
	// and false if not.
	Get(key int) (int, bool)
	// Set updates value associated with the key.
	//
	// If there is no key in the cache new (key, value) pair is created.
	Set(key, value int)
	// Range calls function f on all elements of the cache
	// in increasing access time order.
	//
	// Stops earlier if f returns false.
	Range(f func(key, value int) bool)
	// Clear removes all keys and values from the cache.
	Clear()
}

type accessRecord struct {
	key  int
	time int
}

type item struct {
	value int
	time  int
}

type lruCache struct {
	cap     int
	curTime int
	items   map[int]item
	dq      *list.List
}

func (c *lruCache) Get(key int) (int, bool) {
	it, ok := c.items[key]
	if !ok {
		return 0, false
	}
	c.curTime++
	it.time = c.curTime
	c.items[key] = it
	c.dq.PushBack(accessRecord{key: key, time: c.curTime})
	return it.value, true
}

func (c *lruCache) Set(key, value int) {
	if it, ok := c.items[key]; ok {
		c.curTime++
		it.value = value
		it.time = c.curTime
		c.items[key] = it
		c.dq.PushBack(accessRecord{key: key, time: c.curTime})
		return
	}
	c.curTime++
	c.items[key] = item{value: value, time: c.curTime}
	c.dq.PushBack(accessRecord{key: key, time: c.curTime})
	c.evictIfNeeded()
}

func (c *lruCache) evictIfNeeded() {
	for len(c.items) > c.cap {
		front := c.dq.Front()
		if front == nil {
			break
		}
		rec := front.Value.(accessRecord)
		if it, ok := c.items[rec.key]; ok && it.time == rec.time {
			delete(c.items, rec.key)
		}
		c.dq.Remove(front)
	}
}

func (c *lruCache) Range(f func(key, value int) bool) {
	seen := make(map[int]bool)
	for e := c.dq.Front(); e != nil; e = e.Next() {
		rec := e.Value.(accessRecord)
		if seen[rec.key] {
			continue
		}
		if it, ok := c.items[rec.key]; ok && it.time == rec.time {
			seen[rec.key] = true
			if !f(rec.key, it.value) {
				return
			}
		}
	}
}

func (c *lruCache) Clear() {
	c.items = make(map[int]item)
	c.dq.Init()
}
