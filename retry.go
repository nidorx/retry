package retry

import (
	"context"
	"math"
	"time"
)

type BackoffStrategy interface {
	Next(attempt int) int // next The current value of the counter and immediately updates it with the next value
}

// FixedBackOffStrategy A BackoffStrategy that pauses for a fixed period of time before continuing.
type FixedBackOffStrategy struct {
	period int
}

func (b *FixedBackOffStrategy) Next(attempt int) int {
	return b.period
}

// ExponentialBackoffStrategy A BackoffStrategy that increases the back off period for each retry attempt in a given set
// using the exponential function.
type ExponentialBackoffStrategy struct {
	initTime float64
	maxTime  float64
	factor   float64
}

func (b *ExponentialBackoffStrategy) Next(attempt int) int {
	return int(math.Min(math.Pow(b.factor, float64(attempt-1))*b.initTime, b.maxTime))
}

type OnError func(ctx context.Context, err error, attempt int, willRetry bool, nextRetry time.Duration)

// Retry retries a function a given number of times until success is obtained.
type Retry struct {
	retries   int
	unlimited bool
	onError   OnError
	Backoff   BackoffStrategy
}

// New initialize new Retry
func New(numberOfRetries int, onError OnError) *Retry {
	strategy := &Retry{onError: onError}

	// default backoff
	strategy.SetFixedBackOff(1000)
	strategy.SetNumberOfRetries(numberOfRetries)

	return strategy
}

// SetNumberOfRetries Set the number of retries that are to be attempted before giving up. To try forever, use -1.
func (r *Retry) SetNumberOfRetries(retries int) {
	r.retries = retries
	r.unlimited = retries < 0
}

func (r *Retry) SetFixedBackOff(period int) {
	r.Backoff = &FixedBackOffStrategy{
		period: period,
	}
}

// SetExponentialBackoff
// initTime - in milliseconds for which the execution is suspended after the first attempt
// maxTime - in milliseconds for which the execution can be suspended
// factor - is the base of the power by which the waiting time increases
func (r *Retry) SetExponentialBackoff(initTime int, maxTime int, factor float64) {
	r.Backoff = &ExponentialBackoffStrategy{
		initTime: float64(initTime),
		maxTime:  float64(maxTime),
		factor:   factor,
	}
}

// Execute  Keep retrying a callback with a potentially varying wait on each iteration, until one of the following happens:
// - the callback returns nil
// - the number of retries is exceeded, retuning last error
func (r *Retry) Execute(ctx context.Context, callback func(ctx context.Context, attempt int) error) error {
	attempt := 0
	for {
		// Return immediately if ctx is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		attempt++
		err := callback(ctx, attempt)
		if err == nil {
			break
		}

		if r.unlimited || attempt <= r.retries {

			next := time.Duration(r.Backoff.Next(attempt)) * time.Millisecond

			if r.onError != nil {
				r.onError(ctx, err, attempt, true, next)
			}

			t := time.NewTimer(next)
			select {
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			case <-t.C:
				continue
			}
		} else {
			// the number of retries is exceeded.
			if r.onError != nil {
				r.onError(ctx, err, attempt, false, time.Duration(0))
			}
			return err
		}
	}

	// the callback returns nil
	return nil
}
