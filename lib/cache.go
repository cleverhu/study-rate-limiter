package lib

import (
	"container/list"
	"fmt"
	"log"
	"sync"
	"time"
)

type CacheInterface interface {
	Get(key string) interface{}
	Set(key string, value interface{}, ttl time.Duration)
}

var _ CacheInterface = new(cache)

type cache struct {
	lock    sync.Mutex
	data    map[string]*list.Element
	elist   *list.List
	maxSize int
}

func NewCache(maxSize int) CacheInterface {
	cache := &cache{maxSize: maxSize, data: map[string]*list.Element{}, elist: list.New()}
	cache.clear()
	return cache
}

type cacheData struct {
	key      string
	value    interface{}
	expireAt time.Time
}

func newCacheData(key string, value interface{}, ttl time.Duration) *cacheData {
	if ttl == 0 {
		ttl = time.Hour * 24 * 265
	}
	expireAt := time.Now().Add(ttl)
	return &cacheData{key: key, value: value, expireAt: expireAt}
}

func (c *cache) clear() {
	go func() {
		for {
			c.removeExpired()
			time.Sleep(time.Second)
		}
	}()
}

func (c *cache) removeExpired() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, v := range c.data {
		data := v.Value.(*cacheData)
		if data.expireAt.Before(time.Now()) {
			c.removeItem(v)
		}
	}
}

func (c *cache) Get(key string) interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	value := c.data[key]
	if value != nil {
		data := value.Value.(*cacheData)
		if data.expireAt.Before(time.Now()) {
			c.removeItem(value)
			return nil
		}

		c.elist.MoveToFront(value)
		return data.value
	}
	return nil
}

func (c *cache) Set(key string, value interface{}, ttl time.Duration) {
	fmt.Println(key)
	c.lock.Lock()
	defer c.lock.Unlock()
	data := newCacheData(key, value, ttl)

	val := c.data[key]
	if val != nil {
		val.Value = data
		c.elist.MoveToFront(val)
		c.data[key] = val
	} else {
		ele := c.elist.PushFront(data)
		c.data[key] = ele
		if c.elist.Len() > c.maxSize {
			c.removeOldest()
		}
	}

}

func (c *cache) removeOldest() {
	ele := c.elist.Back()
	if ele == nil {
		return
	}
	c.removeItem(ele)
}

func (c *cache) removeItem(item *list.Element) {
	fmt.Println(item.Value.(*cacheData).key, "removed")
	c.elist.Remove(item)
	delete(c.data, item.Value.(*cacheData).key)
}

func (c *cache) Print() {
	e := c.elist.Front()
	for e != nil {
		log.Println(e.Value)
		e = e.Next()
	}
	log.Println()
	for _, v := range c.data {
		log.Println(v.Value)
	}
	log.Println("====")
}
