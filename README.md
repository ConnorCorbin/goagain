# goagain

Goagain is a simple and flexible way to retry a function that may fail.

-------------------------

- [Requirements](#requirements)
- [Installation](#installation)
  - [Retry until successful](#retry-until-successful)
  - [Retry up to 5 times](#retry-up-to-5-times)
  - [Retry with an early exit](#retry-with-an-early-exit)
  - [Retry with a delay](#retry-with-a-delay)
- [Usage](#usage)
- [License](#license)

-------------------------

## Requirements

Go 1.20 or higher.

-------------------------

## Installation

First, import it using:

```go
import "github.com/ConnorCorbin/goagain"
```

Then, install the latest version using:

```bash
go install github.com/ConnorCorbin/goagain@latest
```

-------------------------

## Usage

### Retry until successful

Retry a function until it succeeds:

```go
doResult, err := Do(
    context.Background(),
    func() error {
        return errors.New("retry until successful")
    },
    nil,
)
if err != nil {
    log.Fatal(err)
}
```

### Retry up to 5 times

Retry a function up to 5 times:

```go
doResult, err := Do(
    context.Background(),
    func() error {
        return errors.New("retry up to 5 times")
    },
    &DoOptions{
        MaxRetries: 5,
    },
)
if err != nil {
    log.Fatal(err)
}
```

### Retry with an early exit

Retry a function up to 5 times but exit early if a certain condition is met:

```go
doResult, err := Do(
    context.Background(),
    func() error {
        return errors.New("retry with an early exit")
    },
    &DoOptions{
        MaxRetries: 5,
        RetryFunc: func(currentResult *DoResult) error {
            isUnrecoverableErr := true
            if isUnrecoverableErr {
                return errors.New("last error is unrecoverable")
            }

            return nil
        },
    },
)
if err != nil {
    log.Fatal(err)
}
```

### Retry with a delay

Retry a function up to 5 times with a delay between attempts:

```go
doResult, err := Do(
    context.Background(),
    func() error {
        return errors.New("retry with a delay")
    },
    &DoOptions{
        MaxRetries: 5,
        DelayFunc: func(currentResult *DoResult) time.Duration {
            return 5 * time.Second
        },
    },
)
if err != nil {
    log.Fatal(err)
}
```

-------------------------

## License

Goagain is licensed under version 2.0 of the [Apache License](LICENSE).
