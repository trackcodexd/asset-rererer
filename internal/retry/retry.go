package retry

import (
	"math"
	"time"
)

func getDelay(o *retryOptions, tries int) float64 {
	delay := o.Delay * (o.BackOff * float64(tries))

	if o.MaxDelay == 0 {
		return delay
	}

	return math.Min(delay, o.MaxDelay)
}

func Do[T any](options *retryOptions, callback func() (T, error)) (T, error) {
	var tries int

	for {
		tries++

		res, err := callback()
		if err == nil {
			return res, nil
		}

		switch err := err.(type) {
		case *ExitRetry:
			return res, err.Err
		case *ContinueRetry:
			if !canRetry(options, tries) {
				return res, err.Err
			}

			<-time.After(time.Duration(getDelay(options, tries) * float64(time.Second)))
		default:
			return res, err
		}
	}
}
