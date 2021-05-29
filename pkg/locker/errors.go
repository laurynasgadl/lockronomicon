package locker

import "errors"

var (
	ErrLockTaken         = errors.New("lock already taken")
	ErrLockNotExist      = errors.New("lock does not exist")
	ErrReadLock          = errors.New("could not read lock info")
	ErrRemoveLock        = errors.New("could not remove lock")
	ErrEncodeMetadata    = errors.New("could not encode metadata")
	ErrDecodeMetadata    = errors.New("could not decode metadata")
	ErrWriteMetadata     = errors.New("could not write metadata")
	ErrReadMetadata      = errors.New("could not read metadata")
	ErrRemoveMetadata    = errors.New("could not remove metadata")
	ErrGenNumberMismatch = errors.New("generation number mismatch")
)
