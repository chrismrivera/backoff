package backoff

import (
	"testing"
	"time"
)

func TestBackoffWaitTime(t *testing.T) {
	type expResult struct {
		a   int
		b   int
		max int
	}

	expResults := []expResult{
		{1, 1, 1},
		{1, 2, 2},
		{2, 3, 3},
		{3, 5, 5},
		{5, 8, 8},
		{8, 13, 13},
		// max wait time is 15 seconds
		{13, 21, 15},
	}

	bw := New(time.Second * 15)

	for _, er := range expResults {
		wt := bw.waitTime()

		if bw.a != er.a {
			t.Fatalf("a was %d, expected %d\n", bw.a, er.a)
		}

		if bw.b != er.b {
			t.Fatalf("b was %d, expected %d\n", bw.b, er.b)
		}

		high := time.Duration(er.max) * time.Second
		low := high / 2

		if wt < low || wt > high {
			t.Fatalf("wait time out of range: %d outside of [%d, %d]", wt, low, high)
		}
	}
}
