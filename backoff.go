package backoff

import (
	"time"
)

type FatalError struct {
	Err error
}

func (fe FatalError) Error() string {
	return fe.Err.Error()
}

type Backoff struct {
	a               int
	b               int
	maxWait         int
	waitCalledCount int
}

func New(maxWait int) *Backoff {
	return &Backoff{0, 1, maxWait, 0}
}

func (bw *Backoff) Wait() {
	bw.waitCalledCount++

	bw.b, bw.a = bw.b+bw.a, bw.b

	wait := bw.b
	if wait > bw.maxWait {
		wait = bw.maxWait
	}

	<-time.After(time.Second * time.Duration(wait))
}

func (bw *Backoff) Reset() {
	bw.a = 0
	bw.b = 1
}

func (bw *Backoff) WaitCalledCount() int {
	return bw.waitCalledCount
}

func (bw *Backoff) Try(attempts int, f func() error) error {
	for {
		err := f()
		if err != nil {
			if fatalErr, ok := err.(FatalError); ok {
				return fatalErr.Err
			}

			if bw.waitCalledCount >= attempts-1 {
				return err
			}

			bw.Wait()
			continue
		}

		return nil
	}
}

func (bw *Backoff) TryWithDeadline(relativeDeadline time.Duration, f func() error) error {
	deadline := time.Now().Add(relativeDeadline)

	for {
		err := f()
		if err != nil {
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

func Try(maxWait, attempts int, f func() error) error {
	bw := New(maxWait)
	return bw.Try(attempts, f)
}

func TryWithDeadline(maxWait int, relativeDeadline time.Duration, f func() error) error {
	bw := New(maxWait)
	return bw.TryWithDeadline(relativeDeadline, f)
}
