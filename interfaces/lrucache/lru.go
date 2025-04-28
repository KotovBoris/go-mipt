//go:build !solution

package lrucache

import "container/list"

func New(cap int) Cache {
	return &lruCache{
		cap:   cap,
		items: make(map[int]item),
		dq:    list.New(),
	}
}
