package pqx

import (
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// IsSerializationFailure returns a boolean indicating whether the error
// is 40001 (serialization_failure).
//
// It tries to unwrap err using github.com/pkg/errors.Cause().
func IsSerializationFailure(err error) bool {
	pqErr, ok := errors.Cause(err).(*pq.Error)
	return ok && pqErr.Code.Name() == "serialization_failure"
}

// Serialize executes given func, which is supposed to run single
// transaction and return (possibly wrapped) error if transaction fail.
//
// It will re-execute given func for up to 10 times in case it fails
// with 40001 (serialization_failure) error.
//
// Returns value returned by last doTx call.
func Serialize(doTx func() error) error {
	const maxTries = 10
	err, try := doTx(), 1
	for IsSerializationFailure(err) && try < maxTries {
		err, try = doTx(), try+1
	}
	return err
}
