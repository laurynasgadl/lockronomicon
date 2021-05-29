package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/laurynasgadl/lockronomicon/pkg/locker"
)

func renderJSON(w http.ResponseWriter, _ *http.Request, data interface{}) (int, error) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}

func renderError(err error) (int, error) {
	var status int

	switch {
	case err == nil:
		status = http.StatusOK
	case errors.Is(err, locker.ErrLockTaken):
		status = http.StatusLocked
	case errors.Is(err, locker.ErrLockNotExist):
		status = http.StatusNotFound
	case errors.Is(err, locker.ErrReadLock):
		status = http.StatusInternalServerError
	case errors.Is(err, locker.ErrRemoveLock):
		status = http.StatusInternalServerError
	case errors.Is(err, locker.ErrEncodeMetadata):
		status = http.StatusInternalServerError
	case errors.Is(err, locker.ErrEncodeMetadata):
		status = http.StatusInternalServerError
	case errors.Is(err, locker.ErrDecodeMetadata):
		status = http.StatusInternalServerError
	case errors.Is(err, locker.ErrWriteMetadata):
		status = http.StatusInternalServerError
	case errors.Is(err, locker.ErrReadMetadata):
		status = http.StatusInternalServerError
	case errors.Is(err, locker.ErrRemoveMetadata):
		status = http.StatusInternalServerError
	case errors.Is(err, locker.ErrGenNumberMismatch):
		status = http.StatusPreconditionFailed
	default:
		status = http.StatusInternalServerError
	}

	return status, err
}
