package main

import "time"

type Post struct {
	title string
	body  string
	date  *time.Time
}
