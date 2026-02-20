package demo

import (
	"fmt"
	"time"
)

// This function is intentionally hard to read to trigger the readability agent.
func ComplexProcess(a int, b string, c []int) (int, error) {
	// Magic numbers and poor naming
	if a > 100 {
		for i := 0; i < len(c); i++ {
			if c[i]%2 == 0 {
				if b == "test" {
					// Deep nesting
					if i > 5 {
						fmt.Println("Processing...")
						time.Sleep(1 * time.Second)
						a = a + c[i]*2
					} else {
						a = a + c[i]
					}
				} else {
					if b == "prod" {
						a = a - c[i]
					}
				}
			} else {
				if a < 1000 {
					a = a * 2
				}
			}
		}
	} else {
		return 0, fmt.Errorf("invalid input")
	}

	// Unclear logic
	x := 0
	for j := 0; j < 10; j++ {
		if j%3 == 0 {
			x += j
		}
	}
	
	if a+x > 5000 {
		return -1, nil
	}

	return a + x, nil
}
