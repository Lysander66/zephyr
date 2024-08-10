package znet

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	defaultMaxRetries  = 3
	defaultWaitTime    = 100 * time.Millisecond
	defaultMaxWaitTime = 2000 * time.Millisecond
)

type (
	// Option is to create convenient retry options like wait time, max retries, etc.
	Option func(*Options)

	// RetryConditionFunc type is for retry condition function
	// input: non-nil Response OR request execution error
	RetryConditionFunc func(*Response, error) bool

	// OnRetryFunc is for side-effecting functions triggered on retry
	OnRetryFunc func(*Response, error)

	// Options struct is used to hold retry settings.
	Options struct {
		maxRetries      int
		waitTime        time.Duration
		maxWaitTime     time.Duration
		retryConditions []RetryConditionFunc
		retryHooks      []OnRetryFunc
	}
)

// Retries sets the max number of retries
func Retries(value int) Option {
	return func(o *Options) {
		o.maxRetries = value
	}
}

// WaitTime sets the default wait time to sleep between requests
func WaitTime(value time.Duration) Option {
	return func(o *Options) {
		o.waitTime = value
	}
}

// MaxWaitTime sets the max wait time to sleep between requests
func MaxWaitTime(value time.Duration) Option {
	return func(o *Options) {
		o.maxWaitTime = value
	}
}

// RetryConditions sets the conditions that will be checked for retry.
func RetryConditions(conditions []RetryConditionFunc) Option {
	return func(o *Options) {
		o.retryConditions = conditions
	}
}

// RetryHooks sets the hooks that will be executed after each retry
func RetryHooks(hooks []OnRetryFunc) Option {
	return func(o *Options) {
		o.retryHooks = hooks
	}
}

// Backoff retries with increasing timeout duration up until X amount of retries
// (Default is 3 attempts, Override with option Retries(n))
func Backoff(operation func() (*Response, error), options ...Option) error {
	// Defaults
	opts := Options{
		maxRetries:      defaultMaxRetries,
		waitTime:        defaultWaitTime,
		maxWaitTime:     defaultMaxWaitTime,
		retryConditions: []RetryConditionFunc{},
	}

	for _, o := range options {
		o(&opts)
	}

	var (
		resp *Response
		err  error
	)

	for attempt := 0; attempt <= opts.maxRetries; attempt++ {
		resp, err = operation()
		ctx := context.Background()
		if resp != nil && resp.Request.ctx != nil {
			ctx = resp.Request.ctx
		}
		if ctx.Err() != nil {
			return err
		}

		needsRetry := err != nil // retry on a few operation errors by default

		for _, condition := range opts.retryConditions {
			needsRetry = condition(resp, err)
			if needsRetry {
				break
			}
		}

		if !needsRetry {
			return err
		}

		for _, hook := range opts.retryHooks {
			hook(resp, err)
		}

		// Don't need to wait when no retries left.
		// Still run retry hooks even on last retry to keep compatibility.
		if attempt == opts.maxRetries {
			return err
		}

		waitTime, err2 := sleepDuration(opts.waitTime, opts.maxWaitTime, attempt)
		if err2 != nil {
			if err == nil {
				err = err2
			}
			return err
		}

		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

func sleepDuration(min, max time.Duration, attempt int) (time.Duration, error) {
	const maxInt = 1<<31 - 1 // max int for arch 386
	if max < 0 {
		max = maxInt
	}
	return jitterBackoff(min, max, attempt), nil
}

// Return capped exponential backoff with jitter
// http://www.awsarchitectureblog.com/2015/03/backoff.html
func jitterBackoff(min, max time.Duration, attempt int) time.Duration {
	base := float64(min)
	capLevel := float64(max)

	temp := math.Min(capLevel, base*math.Exp2(float64(attempt)))
	ri := time.Duration(temp / 2)
	if ri == 0 {
		ri = time.Nanosecond
	}
	result := randDuration(ri)

	if result < min {
		result = min
	}

	return result
}

var rnd = newRnd()
var rndMu sync.Mutex

func randDuration(center time.Duration) time.Duration {
	rndMu.Lock()
	defer rndMu.Unlock()

	var ri = int64(center)
	var jitter = rnd.Int63n(ri)
	return time.Duration(math.Abs(float64(ri + jitter)))
}

func newRnd() *rand.Rand {
	var seed = time.Now().UnixNano()
	var src = rand.NewSource(seed)
	return rand.New(src)
}
