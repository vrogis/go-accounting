package accounting

import (
	"container/list"
	"github.com/vrogis/go-event"
	"sync"
	"time"
)

type Take[TValue valueConstraint] struct {
	mtx       sync.Mutex
	amount    *Amount[TValue]
	element   *list.Element
	fullEvent event.Event[TValue]
	want      TValue
	taken     TValue
}

func (take *Take[TValue]) IsActive() bool {
	take.mtx.Lock()
	defer take.mtx.Unlock()

	return take.isActive()
}

func (take *Take[TValue]) IsFull() bool {
	take.mtx.Lock()
	defer take.mtx.Unlock()

	return take.isFull()
}

func (take *Take[TValue]) Want() TValue {
	return take.want
}

func (take *Take[TValue]) Left() TValue {
	take.mtx.Lock()
	defer take.mtx.Unlock()

	return take.want - take.taken
}

func (take *Take[TValue]) Taken() TValue {
	take.mtx.Lock()
	defer take.mtx.Unlock()

	return take.taken
}

func (take *Take[TValue]) OnFull(subscriber event.Subscriber[TValue]) {
	take.mtx.Lock()

	if take.isActive() {
		take.fullEvent.On(subscriber)

		take.mtx.Unlock()

		return
	}

	if take.isFull() {
		go subscriber(take.taken)
	}

	take.mtx.Unlock()
}

func (take *Take[TValue]) WaitChan() <-chan struct{} {
	waitChan := make(chan struct{})

	take.mtx.Lock()

	if !take.isActive() || take.isFull() {
		take.mtx.Unlock()

		close(waitChan)

		return waitChan
	}

	take.fullEvent.On(func(_ TValue) {
		close(waitChan)
	})

	take.mtx.Unlock()

	return waitChan
}

func (take *Take[TValue]) Wait(waitFor time.Duration, onWaiting func(), interval time.Duration) {
	take.mtx.Lock()

	if !take.isActive() || take.isFull() {
		take.mtx.Unlock()

		return
	}

	take.mtx.Unlock()

	timeChan := time.After(waitFor)
	waitChan := take.WaitChan()

	for {
		select {
		case <-timeChan:
			return
		case <-waitChan:
			return
		default:
			onWaiting()
		}

		time.Sleep(interval)
	}
}

func (take *Take[TValue]) put(amount TValue) (taken TValue, full bool) {
	left := take.want - take.taken - amount

	if left > 0 {
		take.taken += amount

		return amount, false
	}

	taken = amount + left

	take.taken = take.want

	return taken, true
}

func makeWaitingTake[TValue valueConstraint](amount *Amount[TValue], taken TValue, want TValue) *Take[TValue] {
	return &Take[TValue]{
		amount: amount,
		taken:  taken,
		want:   want,
	}
}

func makeSuccessTake[TValue valueConstraint](amount *Amount[TValue], want TValue) *Take[TValue] {
	return &Take[TValue]{
		amount: amount,
		want:   want,
		taken:  want,
	}
}

func (take *Take[TValue]) isActive() bool {
	return nil != take.element
}

func (take *Take[TValue]) isFull() bool {
	return take.want == take.taken
}

func (take *Take[TValue]) finish() {
	take.element = nil
}
