package mutexkv

import (
	"log"
	"sync"
)

// MutexKV is a simple key/value store for arbitrary mutexes. It must be used
// when creating resources, since some of them might restart the servers.
// Not using MutexKV might lead to inconsistances because some resources
// might reace each other.
type MutexKV struct {
	lock  sync.Mutex
	store map[string]*sync.Mutex
}

// Lock the mutex for the given key. The caller is responsible for calling
// Unlock for the same key
func (m *MutexKV) Lock(key string) {
	log.Printf("[DEBUG] Locking %s", key)
	m.get(key).Lock()
	log.Printf("[DEBUG] Locked %s", key)
}

// Unlock the mutex for the given key. Caller must have called Lock for the
// same key first.
func (m *MutexKV) Unlock(key string) {
	log.Printf("[DEBUG] Unlocking %s", key)
	m.get(key).Unlock()
	log.Printf("[DEBUG] Unlocked %s", key)
}

// Returns a mutex for the given key, no guarantee of its lock status
func (m *MutexKV) get(key string) *sync.Mutex {
	m.lock.Lock()
	defer m.lock.Unlock()
	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.store[key] = mutex
	}
	return mutex
}

// NewMutexKV returns a properly initialized MutexKV
func NewMutexKV() *MutexKV {
	return &MutexKV{
		store: make(map[string]*sync.Mutex),
	}
}
