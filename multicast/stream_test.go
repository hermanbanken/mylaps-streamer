package multicast

import (
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestMulticastStream(t *testing.T) {

	var input = make(chan Event)
	var ms = MakeMulticastStream(input)
	var wg sync.WaitGroup
	wg.Add(1)

	var i = 0
	var countA = 5  // sync
	var countB = 10 // async

	// Pipe some initial events
	for ; i < countA; i++ {
		var e = Event{time.Now().Add(time.Duration(i) * time.Second), i}
		input <- e
	}
	time.Sleep(1 * time.Millisecond)

	var arr = []byte{}
	go func() {
		for count := range ms.Listeners {
			arr = append([]byte(arr), byte(count)+byte('0'))
		}
	}()

	// Pipe 10 more async
	go func(i int) {
		time.Sleep(10 * time.Millisecond)
		for end := i + countB; i < end; i++ {
			var e = Event{time.Now().Add(time.Duration(i) * time.Second), i}
			input <- e
			time.Sleep(1 * time.Millisecond)
		}
		close(input)
		wg.Done()
	}(i)

	// Start some workers
	worker := func(i int, replay bool, unsubAfter int, expectedReceive int) {
		var received = 0
		if replay {
			var events, unsub = ms.LiveWithReplay(func(e Event, i, j int) bool { return true })
			for e := range events {
				log.Debugf("Worker %d got %v\n", i, e.Data)
				received++
				if unsubAfter > 0 && received >= unsubAfter {
					unsub <- 1
				}
			}
		} else {
			var events, unsub = ms.Live()
			for e := range events {
				log.Debugf("Worker %d got %v\n", i, e.Data)
				received++
				if unsubAfter >= 0 && received >= unsubAfter {
					unsub <- 1
				}
			}
		}
		if received != expectedReceive {
			t.Errorf("Expected worker %d to receive %d events but it got %d", i, expectedReceive, received)
		}
	}

	go worker(1, false, -1, countB)
	go worker(2, true, -1, countA+countB)
	go worker(3, false, 3, 3)

	wg.Wait()

	if string(arr) != "1232" {
		t.Errorf("Expected listener count to go to 3 and go back to 2.")
	}
}
