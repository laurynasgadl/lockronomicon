package locker

import "time"

type Locker interface {
	// Lock accepts a lock key as well as the TTL for the lock
	// and returns the generation number if lock was acquired or
	// an error otherwise
	Lock(key string, ttl time.Duration) (int64, error)

	// Refresh accepts a lock key as well as a generation number
	// and returns a new generation number if refresh was succesfull
	// or an error otherwise
	Refresh(key string, generation int64) (int64, error)

	// Release accepts a lock key as well as a generation number and
	// returns an error if lock release fails
	Release(key string, generation int64) error

	// Check if lock is expired and returns generation number that
	// should be used for lock release if it is expired
	Expired(key string) (int64, bool, error)
}

// compile time check to ensure interface implementation
var _ Locker = &FsLocker{}
