package accounting

import (
	"container/list"
	"github.com/vrogis/go-event"
	"sync"
	"time"
)

type Take[TValue valueConstraint] struct {
	mtx          sync.Mutex
	amount       *Amount[TValue]
	element      *list.Element
	successEvent event.Event[TValue]
	want         TValue
	taken        TValue
	success      bool
}

func (take *Take[TValue]) Want() TValue {
	take.mtx.Lock()
	defer take.mtx.Unlock()

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

func (take *Take[TValue]) OnSuccess(success event.Subscriber[TValue]) {
	take.mtx.Lock()

	if !take.hasResult() {
		take.mtx.Unlock()

		take.successEvent.On(success)

		return
	}

	take.mtx.Unlock()

	if take.success {
		success(take.taken)
	}
}

func (take *Take[TValue]) WaitChan() <-chan struct{} {
	finished := make(chan struct{})

	if take.IsFinished() {
		close(finished)

		return finished
	}

	take.OnSuccess(func(_ TValue) {
		close(finished)
	})

	return finished
}

func (take *Take[TValue]) Waiting(waitFor time.Duration, onWaiting func(*Take[TValue]), interval time.Duration) {
	var timeEnd func() bool

	if waitFor >= 0 {
		start := time.Now()
		end := start.Add(waitFor)

		timeEnd = func() bool {
			return start.After(end) || start.Equal(end)
		}
	} else {
		timeEnd = func() bool {
			return false
		}
	}

	for {
		take.mtx.Lock()

		if take.hasResult() {
			take.mtx.Unlock()

			return
		}

		take.mtx.Unlock()

		onWaiting(take)

		if timeEnd() {
			return
		}

		time.Sleep(interval)
	}
}

func (take *Take[TValue]) IsFinished() bool {
	take.mtx.Lock()
	defer take.mtx.Unlock()

	return take.hasResult()
}

func (take *Take[TValue]) IsSuccess() bool {
	take.mtx.Lock()
	defer take.mtx.Unlock()

	return take.success
}

func (take *Take[TValue]) put(amount TValue) (taken TValue, full bool) {
	left := take.want - take.taken - amount

	if left > 0 {
		take.taken += amount

		return amount, false
	}

	taken = amount + left

	take.taken = take.want

	take.finish(true)

	take.successEvent.Trigger(take.taken)

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
		amount:  amount,
		want:    want,
		taken:   want,
		success: true,
	}
}

func (take *Take[TValue]) hasResult() bool {
	return nil == take.element
}

func (take *Take[TValue]) finish(success bool) {
	take.element = nil
	take.success = success
}
