package syncmap

import (
	"encoding/json"
	"sync"
)

func New[A any]() *SyncMap[A] {
	return &SyncMap[A]{data: make(map[string]A)}
}

type SyncMap[A any] struct {
	data map[string]A
	lock sync.RWMutex
}

func (s *SyncMap[A]) Put(key string, value A) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = value
}

func (s *SyncMap[A]) Get(key string) (A, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	val, ok := s.data[key]
	return val, ok
}

func (s *SyncMap[A]) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.data, key)
}

func (s *SyncMap[A]) MarshalJSON() ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return json.Marshal(s.data)
}

func (s *SyncMap[A]) UnmarshalJSON(data []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return json.Unmarshal(data, &s.data)
}

func (s *SyncMap[A]) Keys() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := make([]string, 0, len(s.data))
	for key := range s.data {
		result = append(result, key)
	}
	return result
}

func (s *SyncMap[A]) Values() []A {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := make([]A, 0, len(s.data))
	for _, value := range s.data {
		result = append(result, value)
	}
	return result
}
