package accounting

import (
	"github.com/vrogis/go-lock"
)

type Transaction[TOperation any, TValue valueConstraint] struct {
	changes map[*Amount[TValue]]*transactionChange[TOperation, TValue]
}

type transactionChange[TOperation any, TValue valueConstraint] struct {
	operational Operational[TOperation, TValue]
	operations  []TOperation
	changeFunc  func()
}

func (t *Transaction[TOperation, TValue]) Change(operational Operational[TOperation, TValue], operation TOperation) {
	t.lazyInit()

	if existingChange, ok := t.changes[operational.Amount()]; ok {
		existingChange.operations = append(existingChange.operations, operation)

		return
	}

	newChange := &transactionChange[TOperation, TValue]{
		operational: operational,
		operations:  []TOperation{operation},
		changeFunc: func() {
			operational.Change(operation)
		},
	}

	t.changes[operational.Amount()] = newChange
}

func (t *Transaction[TOperation, TValue]) Commit() {
	t.lazyInit()

	amounts := make([]*Amount[TValue], len(t.changes))
	i := 0

	for amount := range t.changes {
		amounts[i] = amount

		i++
	}

	unlock := lock.Acquire(amounts...)

	for _, change := range t.changes {
		for _, operation := range change.operations {
			change.operational.Change(operation)
		}
	}

	unlock()

	t.reset()
}

func (t *Transaction[TOperation, TValue]) Rollback() {
	t.reset()
}

func (t *Transaction[TOperation, TValue]) lazyInit() {
	if nil == t.changes {
		t.reset()
	}
}

func (t *Transaction[TOperation, TValue]) reset() {
	t.changes = make(map[*Amount[TValue]]*transactionChange[TOperation, TValue])
}
