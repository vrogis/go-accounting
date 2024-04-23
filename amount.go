package accounting

import (
	"container/list"
	"sync"
)

type Amount[TValue valueConstraint] struct {
	sync.Mutex
	value       TValue
	activeTakes list.List
}

type valueConstraint interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

func (a *Amount[TValue]) Amount() *Amount[TValue] {
	return a
}

func (a *Amount[TValue]) Available() TValue {
	return a.value
}

func (a *Amount[TValue]) Full() TValue {
	value := a.value

	for it := a.activeTakes.Front(); it != nil; it = it.Next() {
		take := it.Value.(*Take[TValue])

		take.mtx.Lock()
		value += take.taken
		take.mtx.Unlock()
	}

	return value
}

func (a *Amount[TValue]) Take(value TValue) *Take[TValue] {
	if 0 == value {
		return makeSuccessTake[TValue](a, value)
	}

	if value < 0 {
		value = -value
	}

	newAmount := a.value - value

	if newAmount >= 0 {
		a.value = newAmount

		return makeSuccessTake[TValue](a, value)
	}

	var amountTaken TValue

	if a.value > 0 {
		amountTaken = a.value
		a.value = 0
	}

	take := makeWaitingTake[TValue](a, amountTaken, value)

	take.element = a.activeTakes.PushBack(take)

	return take
}

func (a *Amount[TValue]) CancelTake(take *Take[TValue]) {
	take.mtx.Lock()
	defer take.mtx.Unlock()

	if a != take.amount || !take.isActive() {
		return
	}

	a.value += take.taken
	a.activeTakes.Remove(take.element)

	take.finish()

	return
}

func (a *Amount[TValue]) AcceptTake(take *Take[TValue]) {
	take.mtx.Lock()
	defer take.mtx.Unlock()

	if a != take.amount || !take.isActive() {
		return
	}

	a.activeTakes.Remove(take.element)

	take.finish()

	return
}

func (a *Amount[TValue]) TakeIfEnough(value TValue) (amountTaken TValue) {
	if 0 == value {
		return 0
	}

	if value < 0 {
		value = -value
	}

	newAmount := a.value - value

	if newAmount >= 0 {
		a.value = newAmount

		return value
	}

	return value
}

func (a *Amount[TValue]) TakeAsMuch(value TValue) (amountTaken TValue) {
	if 0 == value {
		return 0
	}

	if value < 0 {
		value = -value
	}

	if a.value <= 0 {
		return 0
	}

	newAmount := a.value - value

	if newAmount < 0 {
		amountTaken = a.value

		a.value -= amountTaken

		return
	}

	a.value = newAmount

	return value
}

func (a *Amount[TValue]) TakeForce(value TValue) TValue {
	if 0 == value {
		return 0
	}

	if value < 0 {
		value = -value
	}

	a.value -= value

	return value
}

func (a *Amount[TValue]) Change(value TValue) {
	if 0 == value {
		return
	}

	if value > 0 {
		a.put(value)

		return
	}

	a.value += value

	return
}

func (a *Amount[TValue]) Put(value TValue) TValue {
	if 0 == value {
		return 0
	}

	if value < 0 {
		value = -value
	}

	a.put(value)

	return value
}

func (a *Amount[TValue]) put(value TValue) TValue {
	if a.activeTakes.Len() == 0 {
		a.value += value

		return value
	}

	amountLeft := value

	if a.value < 0 {
		a.value += value

		if a.value <= 0 {
			return value
		}

		amountLeft = a.value
		a.value = 0
	}

	it := a.activeTakes.Front()

	for it != nil {
		take := it.Value.(*Take[TValue])

		take.mtx.Lock()

		if take.isFull() {
			it = it.Next()

			take.mtx.Unlock()

			continue
		}

		amountTaken, full := take.put(amountLeft)
		take.mtx.Unlock()

		if !full {
			return value
		}

		go take.fullEvent.Trigger(take.taken)

		amountLeft -= amountTaken

		if 0 == amountLeft {
			return value
		}

		it = a.activeTakes.Front()
	}

	a.value += amountLeft

	return value
}
