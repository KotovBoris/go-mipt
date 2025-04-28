//go:build !solution

package rwmutex

// A RWMutex is a reader/writer mutual exclusion lock.
// The lock can be held by an arbitrary number of readers or a single writer.
// The zero value for a RWMutex is an unlocked mutex.
//
// If a goroutine holds a RWMutex for reading and another goroutine might
// call Lock, no goroutine should expect to be able to acquire a read lock
// until the initial read lock is released. In particular, this prohibits
// recursive read locking. This is to ensure that the lock eventually becomes
// available; a blocked Lock call excludes new readers from acquiring the
// lock.
type RWMutex struct {
	writer chan struct{}
	reader chan int
}

// New creates *RWMutex.
func New() *RWMutex {
	rwmutex := &RWMutex{
		writer: make(chan struct{}, 1),
		reader: make(chan int, 1),
	}
	rwmutex.writer <- struct{}{}
	return rwmutex
}

// RLock locks rw for reading.
//
// It should not be used for recursive read locking; a blocked Lock
// call excludes new readers from acquiring the lock. See the
// documentation on the RWMutex type.
func (rw *RWMutex) RLock() {
	select {
	case readersCount := <-rw.reader:
		rw.reader <- readersCount + 1
	case <-rw.writer:
		rw.reader <- 1 // First reader
	}
}

// RUnlock undoes a single RLock call;
// it does not affect other simultaneous readers.
// It is a run-time error if rw is not locked for reading
// on entry to RUnlock.
func (rw *RWMutex) RUnlock() {
	readersCount := <-rw.reader
	if readersCount == 1 {
		rw.writer <- struct{}{}
	} else {
		rw.reader <- readersCount - 1
	}
}

// Lock locks rw for writing.
// If the lock is already locked for reading or writing,
// Lock blocks until the lock is available.
func (rw *RWMutex) Lock() {
	<-rw.writer
}

// Unlock unlocks rw for writing. It is a run-time error if rw is
// not locked for writing on entry to Unlock.
//
// As with Mutexes, a locked RWMutex is not associated with a particular
// goroutine. One goroutine may RLock (Lock) a RWMutex and then
// arrange for another goroutine to RUnlock (Unlock) it.
func (rw *RWMutex) Unlock() {
	select {
	case <-rw.reader:
		panic("Unlock while there are readers")
	case <-rw.writer:
		panic("unlock of unlocked mutex")
	default:
		rw.writer <- struct{}{}
	}
}
