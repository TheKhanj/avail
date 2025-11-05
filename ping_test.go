package main

import (
	"context"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	s, err := NewPing("google", "https://google.com")
	if err != nil {
		t.Fatal(err)
		return
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Second*5)
	defer cancel()

	s.Run(ctx)
}
