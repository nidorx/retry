# Retry

[![ci](https://github.com/nidorx/retry/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/nidorx/retry/actions/workflows/ci.yml)

Simple golang library for retry mechanism

```go
package main

import (
    "context"
    "errors"
    "fmt"
	
    "github.com/nidorx/retry"
)

func main() {

    numOfRetries := 3
    
    logErrors := func(ctx context.Context, err error, attempt int, willRetry bool, nextRetry time.Duration) {
        fmt.Println("Something happened")
        fmt.Println(err)
        if willRetry {
            fmt.Println("Retrying in " + (nextRetry).String() + "...")
        }
    }

    // 1) create a retry strategy (FixedBackOff 1000ms)
    retries := retry.New(numOfRetries, logErrors)
    
    // 2) execute your code
    err := retries.Execute(context.Background(), func(ctx context.Context, attempt int) error {
        
        if attempt <= 2 {
            return errors.New("My Custom Error")
        }
        
        return nil
    })
    
    // 3) check for error
    if err == nil {
        fmt.Println(err)
    }
}
```

## FixedBackOff

```go
retries.SetFixedBackOff(500)

// retry 1 = +500ms
// retry 2 = +500ms
```

## ExponentialBackoff

```go
// initTime - in milliseconds for which the execution is suspended after the first attempt
// maxTime - in milliseconds for which the execution can be suspended
// factor - is the base of the power by which the waiting time increases

initTime := 500
maxTime  := 5000
factor   := 2

retries.SetExponentialBackoff(initTime, maxTime, factor)

// retry 1 = +500ms   = Math.pow(2, 0)*500
// retry 2 = +1000ms  = Math.pow(2, 1)*500
// retry 3 = +2000ms  = Math.pow(2, 2)*500
// retry 4 = +4000ms  = Math.pow(2, 3)*500
// retry 5 = +5000ms  = Math.pow(2, 4)*500 = 8000 > 5000
// retry 6 = +5000ms  = Math.pow(2, 5)*500 = 16000 > 5000
```

## Custom Backoff

```go
type CustomBackoff struct {
}

func (b *CustomBackoff) Next(attempt int) int {
    return 200
}

retries.Backoff = &CustomBackoff{}

```
