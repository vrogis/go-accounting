package accounting

import "fmt"

func ExampleTransaction_Change() {
	var amount1, amount2, amount3 Amount[int]
	var transaction Transaction[int, int]

	transaction.Change(&amount1, 10)
	transaction.Change(&amount1, 10)
	transaction.Change(&amount2, -5)
	transaction.Change(&amount2, 7)
	transaction.Change(&amount3, 50)

	transaction.Commit()

	fmt.Println(amount1.Available())
	fmt.Println(amount2.Available())
	fmt.Println(amount3.Available())

	// Output:
	// 20
	// 2
	// 50
}
