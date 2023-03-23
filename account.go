package accounting

type Account[TValue valueConstraint] interface {
	GetAmount() *Amount[TValue]
}
