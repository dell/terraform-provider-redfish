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
