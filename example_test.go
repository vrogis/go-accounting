package accounting

import (
	"fmt"
)

func ExampleAmount_Take() {
	var amount Amount[int]

	amount.TakeForce(100)

	take := amount.Take(100)

	take.OnFinish(func(taken int) {
		fmt.Println(taken)
	})

	amount.Put(80)

	fmt.Println(take.Left())
	fmt.Println(take.Taken())
	fmt.Println(amount.Available())
	fmt.Println(amount.Full())

	amount.Put(25)

	fmt.Println(amount.Available())
	fmt.Println(amount.Full())
	fmt.Println(take.Taken())

	// Output:
	// 20
	// 80
	// 0
	// 80
	// 100
	// 5
	// 5
	// 100
}
