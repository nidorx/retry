package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

var customErr = errors.New("custom")
var executeFn = func(ctx context.Context, attempt int) error {
	if attempt <= 3 {
		return customErr
	}
	return nil
}

func Test_CancelContext(t *testing.T) {

	countError := 0
	fixedBackOff := 500
	sumNextRetry := int64(0)

	ctx, ctxCancel := context.WithCancel(context.Background())

	retries := New(3, func(ctx context.Context, err error, attempt int, willRetry bool, nextRetry time.Duration) {
		sumNextRetry += nextRetry.Milliseconds()
		countError++
		if attempt > 2 {
			ctxCancel()
		}
	})
	retries.SetFixedBackOff(fixedBackOff)

	err := retries.Execute(ctx, executeFn)

	if err == nil {
		t.Fatalf("Error expected")
	}

	if countError != 3 {
		t.Fatalf("Count error not equal, want: %d, got %d", 3, countError)
	}

	if sumNextRetry != 1500 {
		t.Fatalf("nextRetry time error , want: %d, got %d", 1500, sumNextRetry)
	}
}

func Test_CancelContextInsideCallback(t *testing.T) {

	countError := 0
	fixedBackOff := 500
	sumNextRetry := int64(0)

	ctx, ctxCancel := context.WithCancel(context.Background())

	retries := New(3, func(ctx context.Context, err error, attempt int, willRetry bool, nextRetry time.Duration) {
		sumNextRetry += nextRetry.Milliseconds()
		countError++
	})
	retries.SetFixedBackOff(fixedBackOff)

	err := retries.Execute(ctx, func(ctx context.Context, attempt int) error {
		if attempt == 1 {
			ctxCancel()

			return customErr
		}

		return nil
	})

	if err == nil {
		t.Fatalf("Error expected")
	}

	if countError != 1 {
		t.Fatalf("Count error not equal, want: %d, got %d", 1, countError)
	}

	if sumNextRetry != 500 {
		t.Fatalf("nextRetry time error , want: %d, got %d", 500, sumNextRetry)
	}
}

func Test_FixedBackOffSuccess(t *testing.T) {

	countError := 0
	fixedBackOff := 500
	sumNextRetry := int64(0)

	retries := New(3, func(ctx context.Context, err error, attempt int, willRetry bool, nextRetry time.Duration) {
		sumNextRetry += nextRetry.Milliseconds()
		countError++
	})
	retries.SetFixedBackOff(fixedBackOff)

	err := retries.Execute(context.Background(), executeFn)

	if err != nil {
		t.Fatalf("Error not expected")
	}

	if countError != 3 {
		t.Fatalf("Count error not equal, want: %d, got %d", 3, countError)
	}

	if sumNextRetry != 1500 {
		t.Fatalf("nextRetry time error , want: %d, got %d", 1500, sumNextRetry)
	}
}

func Test_FixedBackOffError(t *testing.T) {

	countError := 0
	fixedBackOff := 500
	sumNextRetry := int64(0)

	retries := New(2, func(ctx context.Context, err error, attempt int, willRetry bool, nextRetry time.Duration) {
		sumNextRetry += nextRetry.Milliseconds()
		countError++
	})
	retries.SetFixedBackOff(fixedBackOff)

	err := retries.Execute(context.Background(), executeFn)

	if err == nil {
		t.Fatalf("Error expected")
	}

	if countError != 3 {
		t.Fatalf("Count error not equal, want: %d, got %d", 3, countError)
	}

	if sumNextRetry != 1000 {
		t.Fatalf("nextRetry time error , want: %d, got %d", 1000, sumNextRetry)
	}
}

func Test_ExponentialBackOffSuccess(t *testing.T) {

	countError := 0
	sumNextRetry := int64(0)

	retries := New(3, func(ctx context.Context, err error, attempt int, willRetry bool, nextRetry time.Duration) {
		sumNextRetry += nextRetry.Milliseconds()
		countError++
	})
	retries.SetExponentialBackoff(500, 5000, 2)

	err := retries.Execute(context.Background(), executeFn)

	if err != nil {
		t.Fatalf("Error not expected")
	}

	if countError != 3 {
		t.Fatalf("Count error not equal, want: %d, got %d", 3, countError)
	}

	if sumNextRetry != 3500 {
		// Math.pow(2, 0)*500 = 500
		// Math.pow(2, 1)*500 = 1000
		// Math.pow(2, 2)*500 = 2000
		// 500 + 1000 + 2000
		t.Fatalf("nextRetry time error , want: %d, got %d", 3500, sumNextRetry)
	}
}

func Test_ExponentialBackOffMaxTime(t *testing.T) {

	countError := 0
	sumNextRetry := int64(0)

	retries := New(3, func(ctx context.Context, err error, attempt int, willRetry bool, nextRetry time.Duration) {
		sumNextRetry += nextRetry.Milliseconds()
		countError++
	})
	retries.SetExponentialBackoff(500, 900, 2)

	err := retries.Execute(context.Background(), executeFn)

	if err != nil {
		t.Fatalf("Error not expected")
	}

	if countError != 3 {
		t.Fatalf("Count error not equal, want: %d, got %d", 3, countError)
	}

	if sumNextRetry != 2300 {
		// Math.pow(2, 0)*500 = 500
		// Math.pow(2, 1)*500 = 1000 > 900
		// Math.pow(2, 2)*500 = 2000 > 900
		// 500 + 900 + 900 = 2300
		t.Fatalf("nextRetry time error , want: %d, got %d", 2300, sumNextRetry)
	}
}

func Test_ExponentialBackOffError(t *testing.T) {

	countError := 0
	sumNextRetry := int64(0)

	retries := New(2, func(ctx context.Context, err error, attempt int, willRetry bool, nextRetry time.Duration) {
		sumNextRetry += nextRetry.Milliseconds()
		countError++
	})
	retries.SetExponentialBackoff(500, 5000, 2)

	err := retries.Execute(context.Background(), executeFn)

	if err == nil {
		t.Fatalf("Error expected")
	}

	if countError != 3 {
		t.Fatalf("Count error not equal, want: %d, got %d", 3, countError)
	}

	if sumNextRetry != 1500 {
		// Math.pow(2, 0)*500 = 500
		// Math.pow(2, 1)*500 = 1000
		// 500 + 1000
		t.Fatalf("nextRetry time error , want: %d, got %d", 1500, sumNextRetry)
	}
}
