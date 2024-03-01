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
	"sync"
	"testing"
)

func TestMutexKV(t *testing.T) {
	var sum1 int = 0
	var sum2 int = 0
	mutex := NewMutexKV()
	t.Run("test Mutex with two users", func(t *testing.T) {
		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			wggo := &sync.WaitGroup{}
			wggo.Add(10)
			for i := 0; i < 10; i++ {
				go func(number int) {
					defer wggo.Done()
					mutex.Lock("test")
					defer mutex.Unlock("test")
					sum1 += number
				}(i)
			}
			wggo.Wait()
		}()

		go func() {
			defer wg.Done()
			wggo := &sync.WaitGroup{}
			wggo.Add(5)
			for i := 5; i < 10; i++ {
				go func(number int) {
					defer wggo.Done()
					mutex.Lock("test2")
					defer mutex.Unlock("test2")
					sum2 += number
				}(i)
			}
			wggo.Wait()
		}()

		wg.Wait()
		assertSum(t, sum1, 45)
		assertSum(t, sum2, 35)
	})

}

func assertSum(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("the sum wasn't done correctly. Got %d, want %d", got, want)
	}
}
