package internal

import (
	"container/list"
	"fmt"

	"github.com/faiface/beep"
)

type SoundCache struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

type soundEntry struct {
	key         string
	soundBuffer *beep.Buffer
}

func NewSoundCache(capacity int) *SoundCache {
	return &SoundCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (c *SoundCache) Get(soundPackName string, soundName string) (*beep.Buffer, bool) {
	key := fmt.Sprintf("%v_%v", soundPackName, soundName)
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		return el.Value.(*soundEntry).soundBuffer, true
	}
	return nil, false
}

func (c *SoundCache) Put(soundPackName string, soundName string, s *beep.Buffer) {
	key := fmt.Sprintf("%v_%v", soundPackName, soundName)
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		el.Value.(*soundEntry).soundBuffer = s
		return
	}
	el := c.order.PushFront(&soundEntry{key, s})
	c.items[key] = el

	if c.order.Len() > c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*soundEntry).key)
		}
	}
}
