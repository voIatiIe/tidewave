package src

import "sync"

type SafeHashMap[T comparable, V any] struct {
	lock sync.Mutex
	map_ map[T]V
}

func (r *SafeHashMap[T, V]) Lock() *SafeHashMap[T, V] {
	r.lock.Lock()
	return r
}

func (r *SafeHashMap[T, V]) Unlock() {
	r.lock.Unlock()
}

func (r *SafeHashMap[T, V]) Get(key T) (V, bool) {
	defer r.Lock().Unlock()

	val, ok := r.map_[key]

	return val, ok
}

func (r *SafeHashMap[T, V]) Exists(key T) bool {
	defer r.Lock().Unlock()

	_, ok := r.map_[key]

	return ok
}

func (r *SafeHashMap[T, V]) Put(key T, val V) {
	defer r.Lock().Unlock()

	r.map_[key] = val
}

func (r *SafeHashMap[T, V]) Delete(key T) {
	defer r.Lock().Unlock()

	delete(r.map_, key)
}
