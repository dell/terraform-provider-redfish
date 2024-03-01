/*
Copyright (c) 2021-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
