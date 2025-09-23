package main

import (
	"RappaMaster/network/HTTP"
	"RappaMaster/schedule"
	Tracker "RappaMaster/tracker"
	"context"
)

func init() {
	ctx := context.Background()
	go Tracker.StartAll(ctx)
	go HTTP.StartAll(ctx)
	go schedule.StartAll(ctx)
}
