//go:build !solution

package keylock

import (
	"sort"
	"sync"
)

type KeyLock struct {
	mu    sync.Mutex
	locks map[string]chan struct{}
}

func New() *KeyLock {
	return &KeyLock{
		locks: make(map[string]chan struct{}),
	}
}

func (kl *KeyLock) LockKeys(keys []string, cancel <-chan struct{}) (canceled bool, unlock func()) {
	uniqueKeys := dedupAndSort(keys)
	acquired := make([]chan struct{}, 0, len(uniqueKeys))
	for _, key := range uniqueKeys {
		lockCh := kl.getLockChan(key)
		select {
		case <-lockCh:
			acquired = append(acquired, lockCh)
		case <-cancel:
			for _, ch := range acquired {
				ch <- struct{}{}
			}
			return true, nil
		}
	}
	unlock = func() {
		for _, ch := range acquired {
			ch <- struct{}{}
		}
	}
	return false, unlock
}

func dedupAndSort(keys []string) []string {
	if len(keys) == 0 {
		return keys
	}
	keysCopy := make([]string, len(keys))
	copy(keysCopy, keys)
	sort.Strings(keysCopy)
	unique := keysCopy[:1]
	for _, key := range keysCopy[1:] {
		if key != unique[len(unique)-1] {
			unique = append(unique, key)
		}
	}
	return unique
}

func (kl *KeyLock) getLockChan(key string) chan struct{} {
	kl.mu.Lock()
	defer kl.mu.Unlock()
	ch, exists := kl.locks[key]
	if !exists {
		ch = make(chan struct{}, 1)
		ch <- struct{}{}
		kl.locks[key] = ch
	}
	return ch
}
