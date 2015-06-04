package backoff

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type FatalError struct {
	Err error
}

func (fe FatalError) Error() string {
	return fe.Err.Error()
}

type Backoff struct {
	a               int
	b               int
	maxWait         time.Duration
	waitCalledCount int
}

func New(maxWait time.Duration) *Backoff {
	return &Backoff{0, 1, maxWait, 0}
}

func (bw *Backoff) Reset() {
	bw.a = 0
	bw.b = 1
}

func (bw *Backoff) WaitCalledCount() int {
	return bw.waitCalledCount
}

func (bw *Backoff) waitTime() time.Duration {
	bw.b, bw.a = bw.b+bw.a, bw.b

	base := time.Second * time.Duration(bw.b)
	if base > bw.maxWait {
		base = bw.maxWait
	}

	jitter := rand.Int63n(int64(base / 2))
	return base/2 + time.Duration(jitter)
}

func (bw *Backoff) Wait() {
	bw.waitCalledCount++
	<-time.After(bw.waitTime())
}

// Waits for the backoff duration or until stop is read.
// Returns true if interrupted by the stop channel.
func (bw *Backoff) InterruptableWait(stop <-chan struct{}) bool {
	select {
	case <-time.After(bw.waitTime()):
		return false
	case <-stop:
		return true
	}
}

func (bw *Backoff) Try(relativeDeadline time.Duration, f func() error) error {
	deadline := time.Now().Add(relativeDeadline)

	for {
		if err := f(); err != nil {
			if fatalErr, ok := err.(FatalError); ok {
				return fatalErr.Err
			}

			if time.Now().After(deadline) {
				return err
			}

			bw.Wait()
			continue
		}

		return nil
	}
}

func Try(maxWait, relativeDeadline time.Duration, f func() error) error {
	bw := New(maxWait)
	return bw.Try(relativeDeadline, f)
}
