package main

import (
	"RappaMaster/helper"
	"fmt"
)

func main() {
	for err := range helper.GlobalServiceHelper.ErrorHandler {
		fmt.Println(err)
	}
}
