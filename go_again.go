// Package goagain provides a simple and flexible way to retry a function that may fail.
//
// Retrying a function can be useful in situations where network errors or other transient
// issues might cause the function to fail temporarily.
//
// The `Do` function retries the provided function until it succeeds or the maximum number
// of attempts is reached. You can customize the `Do` function by passing in an options
// struct that allows for fine-grained control over the retry behavior.
//
// Example 1: Run until the passed function succeeds.
//
//	doResult, err := Do(
//	    context.Background(),
//	    func() error {
//	        return errors.New("retry until success")
//	    },
//	    nil,
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Example 2: Retry up to 5 times.
//
//	doResult, err := Do(
//	    context.Background(),
//	    func() error {
//	        return errors.New("retry until success or until it reaches a maximum of 5 attempts")
//	    },
//	    &DoOptions{
//	        MaxRetries: 5,
//	    },
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// The Do function returns a DoResult struct containing useful information about the retry
// operation, including the number of attempts made, a slice of errors that occurred during
// the execution of the work function and the start and finish time of the operation. This
// information can be used to diagnose and troubleshoot issues, as well as to measure the overall
// performance of the retry mechanism.
package goagain

import (
	"context"
	"errors"
	"time"
)

// DoOptions represents options that can be used to configure a GoAgain function.
type DoOptions struct {
	// MaxRetries is the maximum number of times to retry the function before giving
	// up. If not specified, the function will be retried an infinite number of times
	// until the context is cancelled.
	MaxRetries uint

	// RetryFunc takes the current result of the function being retried and returns
	// an error indicating whether to retry the function. If it returns nil, the
	// function will be retried. If it returns a non-nil error, the function will not
	// be retried and the error will be returned to the caller.
	RetryFunc func(currentResult *DoResult) error

	// DelayFunc takes the current result of the function being retried and returns
	// the duration to wait before retrying the function. If not specified or if it
	// returns a duration less than or equal to zero, the function will be retried
	// immediately.
	DelayFunc func(currentResult *DoResult) time.Duration
}

// DoResult is a result type returned by a GoAgain function.
type DoResult struct {
	// Attempts is the number of attempts made to execute the function. The
	// initial attempt is counted as 1.
	Attempts uint

	// WorkErrors is a slice of errors that occurred during the execution of
	// the function. If no errors occurred, the slice will be empty.
	WorkErrors []error

	// StartedAt is the time at which the first attempt to execute the
	// function was made.
	StartedAt time.Time

	// FinishedAt is the time at which the final attempt to execute the
	// function was made.
	FinishedAt time.Time
}

// ErrMaxRetries is an error returned by a GoAgain function when the maximum number
// of retries has been reached without success.
var ErrMaxRetries = errors.New("goagain: reached maximum retries")

// Do retries the provided work function until is succeeds, the maximum number of
// attempts is reached or is cancelled by the context.
func Do(ctx context.Context, work func() error, options *DoOptions) (*DoResult, error) {
	var result DoResult
	defer func() {
		result.FinishedAt = time.Now()
	}()

	result.StartedAt = time.Now()

	for {
		select {
		case <-ctx.Done():
			return &result, ctx.Err()
		default:
			result.Attempts++

			if err := work(); err != nil {
				result.WorkErrors = append(result.WorkErrors, err)

				if options == nil {
					continue
				}

				if result.Attempts == options.MaxRetries {
					return &result, ErrMaxRetries
				}

				if options.RetryFunc != nil {
					if err := options.RetryFunc(&result); err != nil {
						return &result, err
					}
				}

				if options.DelayFunc != nil {
					if err := delay(ctx, options.DelayFunc(&result)); err != nil {
						return &result, err
					}
				}
			} else {
				return &result, nil
			}
		}
	}
}

func delay(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
