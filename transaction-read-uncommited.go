package accounting

type TransactionReadUncommitted[TValue valueConstraint] struct {
	changes []readUncommittedChange[TValue]
}

func (t *TransactionReadUncommitted[TValue]) Take(account Account[TValue], value TValue) *Take[TValue] {
	t.lazyInit()
	take := account.GetAmount().Take(value)

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: account.GetAmount(),
		take:   take,
	})

	return take
}

func (t *TransactionReadUncommitted[TValue]) TakeIfEnough(account Account[TValue], value TValue) (amountTaken TValue) {
	t.lazyInit()

	amountTaken = account.GetAmount().TakeIfEnough(value)

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: account.GetAmount(),
		value:  -amountTaken,
	})

	return
}

func (t *TransactionReadUncommitted[TValue]) TakeAsMuch(account Account[TValue], value TValue) (amountTaken TValue) {
	t.lazyInit()

	amountTaken = account.GetAmount().TakeAsMuch(value)

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: account.GetAmount(),
		value:  -amountTaken,
	})

	return
}

func (t *TransactionReadUncommitted[TValue]) TakeForce(account Account[TValue], value TValue) (amountTaken TValue) {
	t.lazyInit()

	amountTaken = account.GetAmount().TakeForce(value)

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: account.GetAmount(),
		value:  -amountTaken,
	})

	return
}

func (t *TransactionReadUncommitted[TValue]) Change(account Account[TValue], value TValue) {
	t.lazyInit()

	account.GetAmount().Change(value)

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: account.GetAmount(),
		value:  value,
	})

	return
}

func (t *TransactionReadUncommitted[TValue]) Put(account Account[TValue], value TValue) (amountPut TValue) {
	t.lazyInit()

	amountPut = account.GetAmount().Put(value)

	t.changes = append(t.changes, readUncommittedChange[TValue]{
		amount: account.GetAmount(),
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
			change.amount.change(-change.value)

			continue
		}

		if change.take.FinishAndGetResult() {
			change.amount.Put(change.take.taken)
		}
	}

	t.changes = make([]readUncommittedChange[TValue], 0)
}

type readUncommittedChange[TValue valueConstraint] struct {
	amount *Amount[TValue]
	take   *Take[TValue]
	value  TValue
}

func (t *TransactionReadUncommitted[TValue]) lazyInit() {
	if nil == t.changes {
		t.changes = make([]readUncommittedChange[TValue], 0)
	}
}
