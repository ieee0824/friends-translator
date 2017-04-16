package main

import (
	"fmt"

	"github.com/NOX73/go-neural"
)

func main() {

	n := neural.NewNetwork(9, []int{9, 9, 4})
	// Randomize sypaseses weights
	n.RandomizeSynapses()

	result := n.Calculate([]float64{0, 1, 0, 1, 1, 1, 0, 1, 0})

	fmt.Println(result)
}
