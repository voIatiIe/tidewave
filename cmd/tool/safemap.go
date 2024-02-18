package tool

import "sync"

type SafeHashMap[T comparable, V any] struct {
	lock_ *sync.Mutex
	map_  map[T]V
}

func NewSafeHashMap[T comparable, V any]() SafeHashMap[T, V] {
	return SafeHashMap[T, V]{
		map_:  make(map[T]V),
		lock_: &sync.Mutex{},
	}
}

func (r *SafeHashMap[T, V]) lock() *SafeHashMap[T, V] {
	r.lock_.Lock()
	return r
}

func (r *SafeHashMap[T, V]) unlock() {
	r.lock_.Unlock()
}

func (r *SafeHashMap[T, V]) Get(key T) (V, bool) {
	defer r.lock().unlock()

	val, ok := r.map_[key]

	return val, ok
}

func (r *SafeHashMap[T, V]) Exists(key T) bool {
	defer r.lock().unlock()

	_, ok := r.map_[key]

	return ok
}

func (r *SafeHashMap[T, V]) Put(key T, val V) {
	defer r.lock().unlock()

	r.map_[key] = val
}

func (r *SafeHashMap[T, V]) Delete(key T) {
	defer r.lock().unlock()

	delete(r.map_, key)
}
