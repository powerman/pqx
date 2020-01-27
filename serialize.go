package pqx

import (
	"errors"
	"math/rand"
	"time"

	"github.com/lib/pq"
	pkgerrors "github.com/pkg/errors"
)

// IsSerializationFailure returns a boolean indicating whether the error
// is 40001 (serialization_failure).
//
// It tries to unwrap err using github.com/pkg/errors.Cause() or errors.As().
func IsSerializationFailure(err error) bool {
	pqErr, ok := pkgerrors.Cause(err).(*pq.Error)
	if !ok {
		ok = errors.As(err, &pqErr)
	}
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
		delay := time.Duration(rand.Intn(20)) * time.Millisecond
		time.Sleep(delay)
		err, try = doTx(), try+1
	}
	return err
}
