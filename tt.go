package main

import "fmt"

func main() {
	nums := []byte("ADFs")
	fmt.Println(string(nums))
	nums = append(nums[:0], "Fe"...)
	fmt.Println(string(nums))
}
