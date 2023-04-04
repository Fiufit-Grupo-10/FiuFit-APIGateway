package main

import "testing"

func add1(x int) int {
	return x + 1
}

func TestAdd(t *testing.T) {
	value := 1

	got := add1(value)
	want := 2

	if got != want {
		t.Errorf("got %d want %d given, %d", got, want, value)
	}
}
