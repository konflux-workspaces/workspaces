package poll

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

const envE2EPollTimeout = "E2E_POLL_TIMEOUT"
const envE2EPollStepDuration = "E2E_POLL_STEP_DURATION"

func init() {
	pollTimeout = max(getDurationFromEnv(envE2EPollTimeout, time.Minute), 10*time.Second)
	pollStepDuration = max(getDurationFromEnv(envE2EPollStepDuration, time.Second), time.Second)
	if pollTimeout < pollStepDuration {
		panic(fmt.Sprintf("Poll timeout (%v) longer than step duration (%v)", pollTimeout, pollStepDuration))
	}
}

var pollTimeout time.Duration
var pollStepDuration time.Duration

func getDurationFromEnv(envVar string, defaultDuration time.Duration) time.Duration {
	value := os.Getenv(envVar)
	duration, err := time.ParseDuration(value)
	if err != nil {
		duration = defaultDuration
	}
	return duration
}

// WaitForConditionImmediately runs the function `condition` periodically to poll the
// status of a condition.  It waits for either condition's first return value
// to be `true`, or for a timeout to be hit.
//
// By default, this timeout is 1 minute, and `condition` is checked every
// second.  These can be overridden with the `E2E_POLL_TIMEOUT` and
// `E2E_POLL_STEP_DURATION` environment variables.
func WaitForConditionImmediately(ctx context.Context, condition func(ctx context.Context) (bool, error)) error {
	return wait.PollUntilContextTimeout(ctx, pollStepDuration, pollTimeout, true, condition)
}

// WaitForConditionImmediatelyJoiningErrors runs the function `condition` periodically to poll the
// status of a condition.  It waits for either condition's first return value
// to be `true`, or for a timeout to be hit.
//
// By default, this timeout is 1 minute, and `condition` is checked every
// second.  These can be overridden with the `E2E_POLL_TIMEOUT` and
// `E2E_POLL_STEP_DURATION` environment variables.
//
// The errors returned by the invocations of the `condition` function are collected and
// -if timeout is hit- returned as a Joined error.
func WaitForConditionImmediatelyJoiningErrors(ctx context.Context, condition func(ctx context.Context) (bool, error)) error {
	errs := []error{}
	err := WaitForConditionImmediately(ctx, func(ctx context.Context) (bool, error) {
		v, err := condition(ctx)
		if err != nil {
			errs = append(errs, err)
			return false, nil
		}

		return v, nil
	})
	return errors.Join(append([]error{err}, errs...)...)
}
