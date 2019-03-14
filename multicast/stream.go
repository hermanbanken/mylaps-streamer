package multicast

import (
	"sync"
	"time"
)

// Event is a generic kind of thing
type Event struct {
	T    time.Time
	Data interface{}
}

// MulticastStream is one of a special kind
type MulticastStream struct {
	AllEvents []Event
	LastEvent Event
	lock      *sync.Mutex
	cond      *sync.Cond
	done      chan int
	listeners int
	Listeners chan int
}

// MakeMulticastStream initializes a multicast stream
func MakeMulticastStream(generator <-chan Event) *MulticastStream {
	var lock sync.Mutex
	var generatorDone = make(chan int, 1)
	var listeners = make(chan int)
	var s = MulticastStream{
		AllEvents: make([]Event, 0),
		LastEvent: Event{},
		lock:      &lock,
		cond:      sync.NewCond(&lock),
		done:      generatorDone,
		listeners: 0,
		Listeners: listeners,
	}
	go func() {
		for {
			e, ok := <-generator
			if !ok {
				generatorDone <- 1
				return
			}
			lock.Lock()
			s.LastEvent = Event{e.T, e.Data}
			s.AllEvents = append(s.AllEvents, e)
			s.cond.Broadcast()
			lock.Unlock()
		}
	}()
	return &s
}

func (ms *MulticastStream) live(firstTime func(chan<- Event)) (<-chan Event, chan<- int) {
	var outputChannel = make(chan Event)
	var unsubscribe = make(chan int, 1)
	var isFirstTime = true

	go loopUntil(func() {
		ms.lock.Lock()
		if isFirstTime {
			ms.addListener()
			firstTime(outputChannel)
			isFirstTime = false
		}
		ms.cond.Wait()
		outputChannel <- ms.LastEvent
		ms.lock.Unlock()
	}, ms.done, func() { close(outputChannel) }, unsubscribe, func() {
		ms.lock.Lock()
		ms.removeListener()
		ms.lock.Unlock()
	})

	return outputChannel, unsubscribe
}

// Live hooks into the MulticastStream and receives events via the channel
func (ms *MulticastStream) Live() (<-chan Event, chan<- int) {
	return ms.live(func(_ chan<- Event) {})
}

// LiveWithReplay hooks into the MulticastStream but first pumps any existing event that matches the conditionFn
func (ms *MulticastStream) LiveWithReplay(conditionFn func(event Event, i int, negI int) bool) (<-chan Event, chan<- int) {
	return ms.live(func(outputChannel chan<- Event) {
		// Playback
		for i, e := range ms.AllEvents {
			if conditionFn(e, i, len(ms.AllEvents)-1) {
				outputChannel <- e
			}
		}
	})
}

func (ms *MulticastStream) addListener() {
	ms.listeners++
	ms.Listeners <- ms.listeners

}
func (ms *MulticastStream) removeListener() {
	ms.listeners--
	ms.Listeners <- ms.listeners
}

// Loops until done utility function
func loopUntil(fn func(), close <-chan int, onClose func(), unsub <-chan int, onUnsub func()) {
	for {
		select {
		case <-close:
			onClose()
			return
		case <-unsub:
			onUnsub()
			return
		default:
			fn()
		}
	}
}
