package multicast

import (
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	lock    sync.Mutex
	cond    = sync.NewCond(&lock)
	counter = 0
)

// This test is meant for remembering how multicast can be done!
func TestMultiplePollers(t *testing.T) {

	waitAndPrint := func(poller int) {
		for {
			lock.Lock()
			cond.Wait()
			message := counter
			log.Debugf("poller %d got %d\n", poller, message)
			lock.Unlock()
		}
	}

	go waitAndPrint(1)
	go waitAndPrint(2)
	go waitAndPrint(3)
	go waitAndPrint(4)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		time.Sleep(1 * time.Millisecond)
		for i := 1; i < 10; i++ {
			lock.Lock()
			counter++
			log.Debugf("Publishing %d\n", counter)
			cond.Broadcast()
			lock.Unlock()
			time.Sleep(1 * time.Millisecond)
		}
		wg.Done()
	}()

	wg.Wait()
}
