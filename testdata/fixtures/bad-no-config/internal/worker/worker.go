package worker

import "os"

func Load() string {
	a := os.Getenv("A")
	b := os.Getenv("B")
	c := os.Getenv("C")
	d := os.Getenv("D")
	return a + b + c + d
}
