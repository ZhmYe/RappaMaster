package main

import (
	"RappaMaster/helper"
	"RappaMaster/network/HTTP"
	"RappaMaster/schedule"
	Tracker "RappaMaster/tracker"
	"context"
	"fmt"
)

func init() {
	ctx := context.Background()
	go Tracker.StartAll(ctx)
	go HTTP.StartAll(ctx)
	go schedule.StartAll(ctx)
}

func main() {
	for err := range helper.GlobalServiceHelper.ErrorHandler {
		fmt.Println(err)
	}
}
