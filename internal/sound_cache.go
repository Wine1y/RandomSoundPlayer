package internal

import (
	"container/list"

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
	soundFormat *beep.Format
}

func NewSoundCache(capacity int) *SoundCache {
	return &SoundCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (c *SoundCache) Get(soundPath string) (*beep.Buffer, *beep.Format, bool) {
	if el, ok := c.items[soundPath]; ok {
		c.order.MoveToFront(el)
		entry := el.Value.(*soundEntry)
		return entry.soundBuffer, entry.soundFormat, true
	}
	return nil, nil, false
}

func (c *SoundCache) Put(soundPath string, buffer *beep.Buffer, format *beep.Format) {
	if el, ok := c.items[soundPath]; ok {
		c.order.MoveToFront(el)
		entry := el.Value.(*soundEntry)
		entry.soundBuffer = buffer
		entry.soundFormat = format
		return
	}
	el := c.order.PushFront(&soundEntry{soundPath, buffer, format})
	c.items[soundPath] = el

	if c.order.Len() > c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*soundEntry).key)
		}
	}
}
