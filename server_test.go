package main

import "testing"

func TestUploading(t *testing.T) {
	minion := NewMinion()

	minion.current = testJob()
	minion.save()
}
