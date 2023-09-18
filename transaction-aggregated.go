package accounting

import (
	"github.com/vrogis/go-lock"
)

type TransactionAggregated[TValue valueConstraint] struct {
	changesIdx map[IAmount[TValue]]*aggregatedChange[TValue]
	changes    []*aggregatedChange[TValue]
}

type aggregatedChange[TValue valueConstraint] struct {
	amount     IAmount[TValue]
	value      TValue
	changeFunc func() error
}

func (a *aggregatedChange[TValue]) Lock() {
	a.amount.Lock()
}

func (a *aggregatedChange[TValue]) Unlock() {
	a.amount.Unlock()
}

func (t *TransactionAggregated[TValue]) Change(amount IAmount[TValue], value TValue) {
	t.lazyInit()

	if existingChange, ok := t.changesIdx[amount]; ok {
		existingChange.value += value

		return
	}

	newChange := &aggregatedChange[TValue]{
		amount: amount,
		value:  value,
		changeFunc: func() error {
			amount.Change(value)

			return nil
		},
	}

	t.changesIdx[amount] = newChange
	t.changes = append(t.changes, newChange)
}

func (t *TransactionAggregated[TValue]) Commit() {
	t.lazyInit()

	unlock := lock.Acquire(t.changes...)

	for _, changeMade := range t.changes {
		changeMade.amount.Change(changeMade.value)
	}

	unlock()

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
	t.changesIdx = make(map[IAmount[TValue]]*aggregatedChange[TValue])
	t.changes = make([]*aggregatedChange[TValue], 0)
}
