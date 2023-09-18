package accounting

type TransactionReadUncommitted[TValue valueConstraint] struct {
	changes []readUncommittedChange[TValue]
}

func (t *TransactionReadUncommitted[TValue]) Take(amount IAmount[TValue], value TValue) *Take[TValue] {
	t.lazyInit()

	amount.Lock()
	take := amount.Take(value)
	amount.Unlock()

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: amount,
		take:   take,
	})

	return take
}

func (t *TransactionReadUncommitted[TValue]) TakeIfEnough(amount IAmount[TValue], value TValue) (amountTaken TValue) {
	t.lazyInit()

	amount.Lock()
	amountTaken = amount.TakeIfEnough(value)
	amount.Unlock()

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: amount,
		value:  -amountTaken,
	})

	return
}

func (t *TransactionReadUncommitted[TValue]) TakeAsMuch(amount IAmount[TValue], value TValue) (amountTaken TValue) {
	t.lazyInit()

	amount.Lock()
	amountTaken = amount.TakeAsMuch(value)
	amount.Unlock()

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: amount,
		value:  -amountTaken,
	})

	return
}

func (t *TransactionReadUncommitted[TValue]) TakeForce(amount IAmount[TValue], value TValue) (amountTaken TValue) {
	t.lazyInit()

	amount.Lock()
	amountTaken = amount.TakeForce(value)
	amount.Unlock()

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: amount,
		value:  -amountTaken,
	})

	return
}

func (t *TransactionReadUncommitted[TValue]) Change(amount IAmount[TValue], value TValue) {
	t.lazyInit()

	amount.Lock()
	amount.Change(value)
	amount.Unlock()

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: amount,
		value:  value,
	})

	return
}

func (t *TransactionReadUncommitted[TValue]) Put(amount IAmount[TValue], value TValue) (amountPut TValue) {
	t.lazyInit()

	amount.Lock()
	amountPut = amount.Put(value)
	amount.Unlock()

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: amount,
		value:  amountPut,
	})

	return
}

func (t *TransactionReadUncommitted[TValue]) Commit() {
	t.changes = make([]readUncommittedChange[TValue], 0)
}

func (t *TransactionReadUncommitted[TValue]) Rollback() {
	if nil == t.changes {
		return
	}

	for _, change := range t.changes {
		if nil == change.take {
			change.amount.Lock()
			change.amount.Change(-change.value)
			change.amount.Unlock()

			continue
		}

		if change.take.amount.FinishTake(change.take) {
			change.amount.Put(change.take.taken)
		}
	}

	t.changes = make([]readUncommittedChange[TValue], 0)
}

type readUncommittedChange[TValue valueConstraint] struct {
	amount IAmount[TValue]
	take   *Take[TValue]
	value  TValue
}

func (t *TransactionReadUncommitted[TValue]) lazyInit() {
	if nil == t.changes {
		t.changes = make([]readUncommittedChange[TValue], 0)
	}
}
