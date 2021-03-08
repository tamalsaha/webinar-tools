package main

import (
	"fmt"
	"time"
)

func main() {
	/*
	3/15/2021 15:00:00
	3/11/2021 3:00:00
	*/
	v := "3/15/2021 15:00:00"
	t, err := time.Parse("1/2/2006 15:04:05", v)
	if err != nil {
		panic(err)
	}
	fmt.Println(t)
}
