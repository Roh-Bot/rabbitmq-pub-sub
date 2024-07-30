package utils

import (
	"github.com/Roh-Bot/rabbitmq-pub-sub/internal/config"
	"github.com/Roh-Bot/rabbitmq-pub-sub/pkg/global"
	"github.com/cenkalti/backoff/v4"
	"time"
)

type Retry struct {
	InitialInterval     time.Duration
	RandomizationFactor float64
	Multiplier          float64
	MaxInterval         time.Duration
	// After MaxElapsedTime the Retry returns Stop.
	// It never stops if MaxElapsedTime == 0.
	MaxElapsedTime time.Duration
	Stop           time.Duration
}

type RetryOpts func(*Retry)

// WithInitialInterval sets the initial interval between retries.
func WithInitialInterval(duration time.Duration) RetryOpts {
	return func(ebo *Retry) {
		ebo.InitialInterval = duration
	}
}

// WithRandomizationFactor sets the randomization factor to add jitter to intervals.
func WithRandomizationFactor(randomizationFactor float64) RetryOpts {
	return func(ebo *Retry) {
		ebo.RandomizationFactor = randomizationFactor
	}
}

// WithMultiplier sets the multiplier for increasing the interval after each retry.
func WithMultiplier(multiplier float64) RetryOpts {
	return func(ebo *Retry) {
		ebo.Multiplier = multiplier
	}
}

// WithMaxInterval sets the maximum interval between retries.
func WithMaxInterval(duration time.Duration) RetryOpts {
	return func(ebo *Retry) {
		ebo.MaxInterval = duration
	}
}

// WithMaxElapsedTime sets the maximum total time for retries.
func WithMaxElapsedTime(duration time.Duration) RetryOpts {
	return func(ebo *Retry) {
		ebo.MaxElapsedTime = duration
	}
}

// WithRetryStopDuration sets the duration after which retries should stop.
func WithRetryStopDuration(duration time.Duration) RetryOpts {
	return func(ebo *Retry) {
		ebo.Stop = duration
	}
}

func RetryOperation(operation func() error, retryOpts ...RetryOpts) error {
	// Use the exponential backoff strategy
	expBackOffOpts := func(b *backoff.ExponentialBackOff) {
		b.InitialInterval = time.Second * time.Duration(config.GetConfig().Backoff.InitialInterval)
		b.Multiplier = config.GetConfig().Backoff.Mulltiplier
		b.MaxInterval = time.Second * time.Duration(config.GetConfig().Backoff.MaxInterval)
		b.MaxElapsedTime = time.Second * time.Duration(config.GetConfig().Backoff.MaxElapsedTime)
	}
	if len(retryOpts) != 0 {
		r := new(Retry)
		for _, fn := range retryOpts {
			fn(r)
		}
		expBackOffOpts = func(b *backoff.ExponentialBackOff) {
			b.InitialInterval = r.InitialInterval
			b.RandomizationFactor = r.RandomizationFactor
			b.Multiplier = r.Multiplier
			b.MaxInterval = r.MaxInterval
			b.MaxElapsedTime = r.MaxElapsedTime
			b.Stop = r.Stop
		}
	}

	expBackoff := backoff.NewExponentialBackOff(expBackOffOpts)
	// Execute the operation with retries
	backoff.NewExponentialBackOff()
	return backoff.Retry(operation, backoff.WithContext(expBackoff, global.CancellationContext()))
}
