package question_answers

import (
	"fmt"
	"math/big"
	"os"
)

func powerOfTwo(input string) string {
	// Convert input string to big.Int
	n := new(big.Int)
	_, ok := n.SetString(input, 10)
	if !ok {
		return "invalid input"
	}

	// Compute 2^n using Exp
	base := big.NewInt(2)
	result := new(big.Int).Exp(base, n, nil)

	return result.String()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run power_of_two.go <number>")
		return
	}

	input := os.Args[1]
	output := powerOfTwo(input)
	fmt.Println(output)
}
