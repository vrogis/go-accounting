package accounting

import (
	"container/list"
	"sync"
)

type IAmount[TValue valueConstraint] interface {
	sync.Locker
	Available() TValue
	Full() TValue
	Take(value TValue) *Take[TValue]
	FinishTake(take *Take[TValue]) (success bool)
	TakeIfEnough(value TValue) (amountTaken TValue)
	TakeAsMuch(value TValue) (amountTaken TValue)
	TakeForce(value TValue) TValue
	Change(value TValue)
	Put(value TValue) TValue
}

type Amount[TValue valueConstraint] struct {
	sync.Mutex
	value        TValue
	waitingTakes list.List
}

type valueConstraint interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

func (a *Amount[TValue]) Available() TValue {
	return a.value
}

func (a *Amount[TValue]) Full() TValue {
	value := a.value

	for it := a.waitingTakes.Front(); it != nil; it = it.Next() {
		value += it.Value.(*Take[TValue]).taken
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
	}

	take := makeWaitingTake[TValue](a, amountTaken, value)

	take.element = a.waitingTakes.PushBack(take)

	a.value = 0

	return take
}

func (a *Amount[TValue]) FinishTake(take *Take[TValue]) (success bool) {
	take.mtx.Lock()
	defer take.mtx.Unlock()

	if take.hasResult() {
		return take.success
	}

	element := take.element
	taken := take.taken

	take.finish(false)

	take.taken = 0

	a.waitingTakes.Remove(element)
	a.value += taken

	return false
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

	a.value -= value

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
