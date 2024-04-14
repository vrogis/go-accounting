package accounting

import (
	"fmt"
)

func ExampleAmount_Take() {
	var amount Amount[int]

	amount.TakeForce(100)

	take := amount.Take(100)

	take.OnFull(func(taken int) {
		fmt.Printf("8) take.OnFull: %d\n", taken)
	})

	amount.Put(80)

	fmt.Printf("1) take.Left(): %d\n", take.Left())
	fmt.Printf("2) take.Taken(): %d\n", take.Taken())
	fmt.Printf("3) amount.Available(): %d\n", amount.Available())
	fmt.Printf("4) amount.Full(): %d\n", amount.Full())

	amount.Put(25)

	fmt.Printf("5) amount.Available(): %d\n", amount.Available())
	fmt.Printf("6) amount.Full(): %d\n", amount.Full())
	fmt.Printf("7) take.Taken(): %d\n", take.Taken())

	<-take.WaitChan()

	amount.AcceptTake(take)

	<-take.WaitChan()

	fmt.Printf("9) amount.Available(): %d\n", amount.Available())
	fmt.Printf("10) amount.Full(): %d\n", amount.Full())

	take2 := amount.Take(50)

	fmt.Printf("11) amount.Available(): %d\n", amount.Available())
	fmt.Printf("12) amount.Full(): %d\n", amount.Full())

	amount.Put(95)

	fmt.Printf("13) take2.Left(): %d\n", take2.Left())
	fmt.Printf("14) take2.Taken(): %d\n", take2.Taken())
	fmt.Printf("15) take2.IsActive(): %t\n", take2.IsActive())
	fmt.Printf("16) take2.IsFull(): %t\n", take2.IsFull())
	fmt.Printf("17) take.IsActive(): %t\n", take.IsActive())
	fmt.Printf("18) take.IsFull(): %t\n", take.IsFull())

	wait := make(chan struct{})

	take.OnFull(func(taken int) {
		fmt.Printf("19) take.OnFull: %d\n", taken)
		close(wait)
	})

	<-wait

	// Unordered output:
	// 1) take.Left(): 20
	// 2) take.Taken(): 80
	// 3) amount.Available(): -100
	// 4) amount.Full(): -20
	// 5) amount.Available(): -95
	// 6) amount.Full(): 5
	// 7) take.Taken(): 100
	// 8) take.OnFull: 100
	// 9) amount.Available(): -95
	// 10) amount.Full(): -95
	// 11) amount.Available(): -95
	// 12) amount.Full(): -95
	// 13) take2.Left(): 0
	// 14) take2.Taken(): 50
	// 15) take2.IsActive(): true
	// 16) take2.IsFull(): true
	// 17) take.IsActive(): false
	// 18) take.IsFull(): true
	// 19) take.OnFull: 100
}
