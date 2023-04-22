package goagain_test

import (
	"context"
	"errors"
	"goagain"
	"testing"
	"time"
)

const shortDuration = 1 * time.Second

var errWork = errors.New("work error")
var errEarlyExit = errors.New("early exit")

func TestDo(t *testing.T) {
	t.Run("should have correct DoResult when first attempt is successful", func(tt *testing.T) {
		r, err := goagain.Do(context.TODO(), func() error { return nil }, nil)

		assertErr(tt, err, nil)
		assertAttempts(tt, r.Attempts, 1)
		assertWorkErrs(tt, r.WorkErrors, nil)
		assertStartedAt(tt, r.StartedAt, r.FinishedAt)
		assertFinishedAt(tt, r.FinishedAt, r.StartedAt)
	})

	t.Run("should have correct DoResult when maximum retries is reached", func(tt *testing.T) {
		r, err := goagain.Do(context.TODO(), func() error { return errWork }, &goagain.DoOptions{
			MaxRetries: 5,
		})

		assertErr(tt, err, goagain.ErrMaxRetries)
		assertAttempts(tt, r.Attempts, 5)
		assertWorkErrs(tt, r.WorkErrors, []error{errWork, errWork, errWork, errWork, errWork})
		assertStartedAt(tt, r.StartedAt, r.FinishedAt)
		assertFinishedAt(tt, r.FinishedAt, r.StartedAt)
	})

	t.Run("should have correct DoResult when retry function returns an error", func(tt *testing.T) {
		r, err := goagain.Do(context.TODO(), func() error { return errWork }, &goagain.DoOptions{
			MaxRetries: 5,
			RetryFunc: func(currentResult *goagain.DoResult) error {
				if currentResult.Attempts == 3 {
					return errEarlyExit
				}

				return nil
			},
		})

		assertErr(tt, err, errEarlyExit)
		assertAttempts(tt, r.Attempts, 3)
		assertWorkErrs(tt, r.WorkErrors, []error{errWork, errWork, errWork})
		assertStartedAt(tt, r.StartedAt, r.FinishedAt)
		assertFinishedAt(tt, r.FinishedAt, r.StartedAt)
	})

	t.Run("should have correct DoResult when context is cancelled before the Do function is called", func(tt *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		r, err := goagain.Do(ctx, func() error { return errWork }, nil)

		assertErr(tt, err, ctx.Err())
		assertAttempts(tt, r.Attempts, 0)
		assertWorkErrs(tt, r.WorkErrors, nil)
		assertStartedAt(tt, r.StartedAt, r.FinishedAt)
		assertFinishedAt(tt, r.FinishedAt, r.StartedAt)
	})

	t.Run("should have correct DoResult when context is cancelled during the retry function", func(tt *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		r, err := goagain.Do(ctx, func() error { return errWork }, &goagain.DoOptions{
			RetryFunc: func(currentResult *goagain.DoResult) error {
				if currentResult.Attempts == 3 {
					cancel()
				}

				return nil
			},
		})

		assertErr(tt, err, ctx.Err())
		assertAttempts(tt, r.Attempts, 3)
		assertWorkErrs(tt, r.WorkErrors, []error{errWork, errWork, errWork})
		assertStartedAt(tt, r.StartedAt, r.FinishedAt)
		assertFinishedAt(tt, r.FinishedAt, r.StartedAt)
	})

	t.Run("should have correct DoResult when context is cancelled during the delay function", func(tt *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		r, err := goagain.Do(ctx, func() error { return errWork }, &goagain.DoOptions{
			DelayFunc: func(currentResult *goagain.DoResult) time.Duration {
				if currentResult.Attempts == 3 {
					cancel()
				}

				return shortDuration
			},
		})

		assertErr(tt, err, ctx.Err())
		assertAttempts(tt, r.Attempts, 3)
		assertWorkErrs(tt, r.WorkErrors, []error{errWork, errWork, errWork})
		assertStartedAt(tt, r.StartedAt, r.FinishedAt)
		assertFinishedAt(tt, r.FinishedAt, r.StartedAt)
	})
}

func assertAttempts(t *testing.T, got uint, want uint) {
	if got != want {
		t.Fatalf("unexpected attempts: \ngot: %v\nwant: %v", got, want)
	}
}

func assertWorkErrs(t *testing.T, got []error, want []error) {
	if len(got) != len(want) {
		t.Fatalf("unexpected work errors: \ngot: %v\nwant: %v", got, want)
	}

	for i := range got {
		if !errors.Is(got[i], want[i]) {
			t.Fatalf("unexpected work errors: \ngot: %v\nwant: %v", got, want)
		}
	}
}

func assertStartedAt(t *testing.T, gotStartedAt time.Time, gotFinishedAt time.Time) {
	if gotStartedAt.After(gotFinishedAt) {
		t.Fatalf("unexpected start at: start time is after finish time: \nstart: %v\nfinish: %v", gotStartedAt, gotFinishedAt)
	}
}

func assertFinishedAt(t *testing.T, gotFinishedAt time.Time, gotStartedAt time.Time) {
	if gotFinishedAt.Before(gotStartedAt) {
		t.Fatalf("unexpected finish at: finish time is before start time: \nstart: %v\nfinish: %v", gotStartedAt, gotFinishedAt)
	}
}

func TestDoResult_LastWorkErr(t *testing.T) {
	t.Run("j", func(tt *testing.T) {
		r := goagain.DoResult{}

		gotErr := r.LastWorkErr()

		assertErr(tt, gotErr, nil)
	})

	t.Run("s", func(tt *testing.T) {
		r := goagain.DoResult{
			WorkErrors: []error{errWork, errEarlyExit},
		}

		gotErr := r.LastWorkErr()

		assertErr(tt, gotErr, errEarlyExit)
	})
}

func assertErr(t *testing.T, got error, want error) {
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: \ngot: %v\nwant: %v", got, want)
	}
}
