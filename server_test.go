package main

import "testing"

func TestUploading(t *testing.T) {
	Configure()

	minion := NewMinion()

	minion.current = testJob()
	minion.save()
}
