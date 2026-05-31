package pqx

import (
	"errors"
	"math/rand/v2"
	"time"

	"github.com/lib/pq"
)

// IsSerializationFailure returns a boolean indicating whether the error
// is 40001 (serialization_failure).
//
// It tries to unwrap err using [errors.As].
func IsSerializationFailure(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return pqErr.Code.Name() == "serialization_failure"
}

// Serialize executes given func, which is supposed to run single
// transaction and return (possibly wrapped) error if transaction fail.
//
// It will re-execute given func for up to 10 times in case it fails
// with 40001 (serialization_failure) error.
//
// Returns value returned by last doTx call.
func Serialize(doTx func() error) error {
	const (
		maxTries         = 10
		maxDelayInMillis = 20
	)
	err, try := doTx(), 1
	for IsSerializationFailure(err) && try < maxTries {
		delay := time.Duration(rand.IntN(maxDelayInMillis)) * time.Millisecond //nolint:gosec // No need in crypto/rand..
		time.Sleep(delay)
		err, try = doTx(), try+1
	}
	return err
}
