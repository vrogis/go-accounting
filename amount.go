package accounting

import (
	"container/list"
	"sync"
)

type Amount[TValue valueConstraint] struct {
	sync.Mutex
	value        TValue
	waitingTakes list.List
}

type valueConstraint interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

func (a *Amount[TValue]) Value() TValue {
	a.Lock()

	value := a.value

	a.Unlock()

	return value
}

func (a *Amount[TValue]) Take(value TValue) *Take[TValue] {
	if 0 == value {
		return makeSuccessTake[TValue](a, value)
	}

	if value < 0 {
		value = -value
	}

	a.Lock()

	newAmount := a.value - value

	if newAmount >= 0 {
		a.value = newAmount

		a.Unlock()

		return makeSuccessTake(a, value)
	}

	var amountWant TValue

	if a.value <= 0 {
		amountWant = value
	} else {
		amountWant = -newAmount
	}

	take := makeWaitingTake(a, amountWant)

	take.element = a.waitingTakes.PushBack(take)

	a.value = 0

	a.Unlock()

	return take
}

func (a *Amount[TValue]) TakeIfEnough(value TValue) (amountTaken TValue) {
	if 0 == value {
		return 0
	}

	if value < 0 {
		value = -value
	}

	a.Lock()

	newAmount := a.value - value

	if newAmount >= 0 {
		a.value = newAmount

		a.Unlock()

		return value
	}

	a.Unlock()

	return value
}

func (a *Amount[TValue]) TakeAsMuch(value TValue) (amountTaken TValue) {
	if 0 == value {
		return 0
	}

	if value < 0 {
		value = -value
	}

	a.Lock()

	if a.value <= 0 {
		return 0
	}

	newAmount := a.value - value

	if newAmount < 0 {
		amountTaken = a.value

		a.value -= amountTaken

		a.Unlock()

		return
	}

	a.value -= value

	a.Unlock()

	return value
}

func (a *Amount[TValue]) TakeForce(value TValue) TValue {
	if 0 == value {
		return 0
	}

	if value < 0 {
		value = -value
	}

	a.Lock()

	a.value -= value

	a.Unlock()

	return value
}

func (a *Amount[TValue]) Change(value TValue) {
	a.Lock()

	a.change(value)

	a.Unlock()
}

func (a *Amount[TValue]) Put(value TValue) TValue {
	if 0 == value {
		return 0
	}

	if value < 0 {
		value = -value
	}

	a.Lock()

	a.put(value)

	a.Unlock()

	return value
}

func (a *Amount[TValue]) change(value TValue) TValue {
	if 0 == value {
		return 0
	}

	if value > 0 {
		a.put(value)

		return value
	}

	a.value += value

	return value
}

func (a *Amount[TValue]) put(value TValue) TValue {
	if a.waitingTakes.Len() == 0 {
		a.value += value

		return value
	}

	it := a.waitingTakes.Front()
	amountLeft := value

	for it != nil {
		take := it.Value.(*Take[TValue])

		take.mtx.Lock()

		if take.hasResult() {
			take.mtx.Unlock()

			it = it.Next()

			continue
		}

		element := take.element
		amountTaken, full := take.put(amountLeft)

		if !full {
			take.mtx.Unlock()

			return value
		}

		amountLeft -= amountTaken

		a.waitingTakes.Remove(element)

		take.mtx.Unlock()

		if 0 == amountLeft {
			return value
		}

		it = a.waitingTakes.Front()
	}

	a.value += amountLeft

	return value
}
