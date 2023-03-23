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
	result       bool
}

func makeWaitingTake[TValue valueConstraint](amount *Amount[TValue], want TValue) *Take[TValue] {
	return &Take[TValue]{
		amount: amount,
		taken:  amount.value,
		want:   want,
	}
}

func makeSuccessTake[TValue valueConstraint](amount *Amount[TValue], amountWant TValue) *Take[TValue] {
	return &Take[TValue]{
		amount: amount,
		want:   amountWant,
		taken:  amountWant,
		result: true,
	}
}

func (take *Take[TValue]) Amount() *Amount[TValue] {
	return take.amount
}

func (take *Take[TValue]) Left() TValue {
	take.mtx.Lock()

	amountLeft := take.want - take.taken

	take.mtx.Unlock()

	return amountLeft
}

func (take *Take[TValue]) Taken() TValue {
	take.mtx.Lock()

	taken := take.taken

	take.mtx.Unlock()

	return taken
}

func (take *Take[TValue]) OnSuccess(success event.Subscriber[TValue]) {
	take.mtx.Lock()

	if !take.hasResult() {
		take.mtx.Unlock()

		take.successEvent.On(success)

		return
	}

	if take.result {
		take.mtx.Unlock()

		success(take.taken)
	}
}

func (take *Take[TValue]) OnWaiting(onWaiting func(*Take[TValue]), interval time.Duration) {
	go func() {
		for {
			take.mtx.Lock()

			if take.hasResult() {
				take.mtx.Unlock()

				return
			}

			take.mtx.Unlock()

			onWaiting(take)

			time.Sleep(interval)
		}
	}()
}

func (take *Take[TValue]) FinishAndGetResult() (result bool) {
	take.mtx.Lock()

	if take.hasResult() {
		defer take.mtx.Unlock()

		return take.result
	}

	element := take.element
	taken := take.taken

	take.setResult(false)

	take.taken = 0

	take.mtx.Unlock()

	take.amount.Lock()

	take.amount.waitingTakes.Remove(element)
	take.amount.value += taken

	take.amount.Unlock()

	return false
}

func (take *Take[TValue]) IsFinished() bool {
	take.mtx.Lock()

	hasResult := take.hasResult()

	take.mtx.Unlock()

	return hasResult
}

func (take *Take[TValue]) IsSuccess() bool {
	take.mtx.Lock()

	result := take.result

	take.mtx.Unlock()

	return result
}

func (take *Take[TValue]) put(amount TValue) (taken TValue, full bool) {
	left := take.want - take.taken - amount

	if left > 0 {
		take.taken += amount

		return amount, false
	}

	taken = amount + left

	take.taken = take.want

	take.setResult(true)

	take.successEvent.Trigger(take.taken)

	return taken, true
}

func (take *Take[TValue]) hasResult() bool {
	return nil == take.element
}

func (take *Take[TValue]) setResult(result bool) {
	take.element = nil
	take.result = result
}
