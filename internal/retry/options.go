package retry

type retryOptions struct {
	MaxTries int
	Delay    float64
	MaxDelay float64
	BackOff  float64
}

func NewOptions(options ...func(*retryOptions)) *retryOptions {
	o := &retryOptions{
		MaxTries: -1,
		Delay:    1,
		MaxDelay: 0,
		BackOff:  1,
	}

	for _, option := range options {
		option(o)
	}

	return o
}

func Tries(tries int) func(*retryOptions) {
	return func(o *retryOptions) {
		o.MaxTries = tries
	}
}

func Delay(delay float64) func(*retryOptions) {
	return func(o *retryOptions) {
		o.Delay = delay
	}
}

func MaxDelay(maxDelay float64) func(*retryOptions) {
	return func(o *retryOptions) {
		o.MaxDelay = maxDelay
	}
}

func BackOff(backOff float64) func(*retryOptions) {
	return func(o *retryOptions) {
		o.BackOff = backOff
	}
}

func canRetry(o *retryOptions, tries int) bool {
	if tries == -1 || tries < o.MaxTries {
		return true
	}
	return false
}
