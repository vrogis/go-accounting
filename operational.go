package accounting

type Operational[TOperation any, TValue valueConstraint] interface {
	Available() TValue
	Full() TValue
	Take(operation TOperation) *Take[TValue]
	TakeIfEnough(operation TOperation) (amountTaken TValue)
	TakeAsMuch(operation TOperation) (amountTaken TValue)
	TakeForce(operation TOperation) TValue
	Change(operation TOperation)
	Put(operation TOperation) TValue
	Amount() *Amount[TValue]
}
