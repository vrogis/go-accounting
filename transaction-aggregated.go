package accounting

import (
	"sort"
	"unsafe"
)

type TransactionAggregated[TValue valueConstraint] struct {
	changesIdx map[*Amount[TValue]]*aggregatedChange[TValue]
	changes    []*aggregatedChange[TValue]
}

type aggregatedChange[TValue valueConstraint] struct {
	amount     *Amount[TValue]
	amountPtr  uintptr
	value      TValue
	changeFunc func() error
}

func (t *TransactionAggregated[TValue]) Change(account Account[TValue], value TValue) {
	t.lazyInit()

	if existingChange, ok := t.changesIdx[account.GetAmount()]; ok {
		existingChange.value += value

		return
	}

	amount := account.GetAmount()

	newChange := &aggregatedChange[TValue]{
		amount:    amount,
		amountPtr: uintptr(unsafe.Pointer(account.GetAmount())),
		value:     value,
		changeFunc: func() error {
			amount.change(value)

			return nil
		},
	}

	t.changesIdx[account.GetAmount()] = newChange
	t.changes = append(t.changes, newChange)
}

func (t *TransactionAggregated[TValue]) Commit() {
	t.lazyInit()

	sort.Slice(t.changes, func(i, j int) bool {
		return t.changes[i].amountPtr < t.changes[j].amountPtr
	})

	for _, changeMade := range t.changes {
		changeMade.amount.Lock()
	}

	for _, changeMade := range t.changes {
		changeMade.amount.change(changeMade.value)
	}

	for _, changeMade := range t.changes {
		changeMade.amount.Unlock()
	}

	t.reset()
}

func (t *TransactionAggregated[TValue]) Rollback() {
	t.reset()
}

func (t *TransactionAggregated[TValue]) lazyInit() {
	if nil == t.changesIdx {
		t.reset()
	}
}

func (t *TransactionAggregated[TValue]) reset() {
	t.changesIdx = make(map[*Amount[TValue]]*aggregatedChange[TValue])
	t.changes = make([]*aggregatedChange[TValue], 0)
}
