package locker

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const rootLockDir = "/tmp/test/locker"

func execFsTest(t *testing.T, fn func(l *FsLocker)) {
	locker, err := NewFsLocker(rootLockDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	defer os.RemoveAll(rootLockDir)

	fn(locker)
}

func TestLockCreatesRequiredFiles(t *testing.T) {
	execFsTest(t, func(l *FsLocker) {
		key := "test.key"

		_, err := l.Lock(key, 100*time.Second)
		if err != nil {
			t.Errorf("fs locker lock unexpected error: %v", err)
		}

		_, err = os.Stat(filepath.Join(rootLockDir, key, "metadata"))
		if err != nil {
			t.Errorf("fs locker unexpected error: %v", err)
		}
	})

}

func TestLockPreventsConsecutiveLock(t *testing.T) {
	execFsTest(t, func(l *FsLocker) {
		key := "test.key"

		_, err := l.Lock(key, 100*time.Second)
		if err != nil {
			t.Errorf("fs locker lock unexpected error: %v", err)
		}

		_, err = l.Lock(key, 100*time.Second)
		if !errors.Is(err, ErrLockTaken) {
			t.Errorf("fs locker expected lock taken error")
		}
	})
}

func TestReleaseAllowsConsecutiveLocks(t *testing.T) {
	execFsTest(t, func(l *FsLocker) {
		key := "test.key"

		gn, err := l.Lock(key, 100*time.Second)
		if err != nil {
			t.Errorf("fs locker lock unexpected error: %v", err)
		}

		err = l.Release(key, gn)
		if err != nil {
			t.Errorf("fs locker release unexpected error: %v", err)
		}

		_, err = l.Lock(key, 100*time.Second)
		if err != nil {
			t.Errorf("fs locker lock unexpected error: %v", err)
		}
	})
}

func TestReleaseFailsOnGenerationNumberMismatch(t *testing.T) {
	execFsTest(t, func(l *FsLocker) {
		key := "test.key"

		gn, err := l.Lock(key, 100*time.Second)
		if err != nil {
			t.Errorf("fs locker lock unexpected error: %v", err)
		}

		err = l.Release(key, gn+10)
		if !errors.Is(err, ErrGenNumberMismatch) {
			t.Errorf("fs locker expected generation number mismatch error")
		}
	})
}

func TestRefreshReturnsValidNewGenerationNumber(t *testing.T) {
	execFsTest(t, func(l *FsLocker) {
		key := "test.key"

		gn, err := l.Lock(key, 100*time.Second)
		if err != nil {
			t.Errorf("fs locker lock unexpected error: %v", err)
		}

		gn2, err := l.Refresh(key, gn)
		if err != nil {
			t.Errorf("fs locker refresh unexpected error: %v", err)
		}

		if gn == gn2 {
			t.Errorf("fs locker refresh returned same generation number: %d", gn)
		}

		err = l.Release(key, gn2)
		if err != nil {
			t.Errorf("fs locker release unexpected error: %v", err)
		}
	})
}

func TestRefreshUpdatesMetadata(t *testing.T) {
	execFsTest(t, func(l *FsLocker) {
		key := "test.key"

		gn, err := l.Lock(key, 100*time.Second)
		if err != nil {
			t.Errorf("fs locker lock unexpected error: %v", err)
		}

		mdFilePath := filepath.Join(l.rootDir, key, metadataFilename)

		mdInfo, err := os.ReadFile(mdFilePath)
		if err != nil {
			t.Errorf("fs locker metadata read unexpected error: %v", err)
		}

		ogMetadata, err := ParseMetadata(mdInfo)
		if err != nil {
			t.Errorf("fs locker metadata parse unexpected error: %v", err)
		}

		time.Sleep(1 * time.Second)
		_, err = l.Refresh(key, gn)
		if err != nil {
			t.Errorf("fs locker refresh unexpected error: %v", err)
		}

		mdInfo, err = os.ReadFile(mdFilePath)
		if err != nil {
			t.Errorf("fs locker metadata read unexpected error: %v", err)
		}

		updatedMetadata, err := ParseMetadata(mdInfo)
		if err != nil {
			t.Errorf("fs locker metadata parse unexpected error: %v", err)
		}

		if ogMetadata.Expires == updatedMetadata.Expires {
			t.Errorf("fs locker metadata expiry not updated: %v", updatedMetadata)
		}

		if ogMetadata.TTL != updatedMetadata.TTL {
			t.Errorf("fs locker metadata TTL unexpectedly updated: %d", updatedMetadata.TTL)
		}
	})
}

func TestRecognizesExpiredLock(t *testing.T) {
	execFsTest(t, func(l *FsLocker) {
		key := "test.key"

		_, err := l.Lock(key, 0)
		if err != nil {
			t.Errorf("fs locker lock unexpected error: %v", err)
		}

		time.Sleep(1 * time.Second)

		_, exp, err := l.Expired(key)
		if err != nil {
			t.Errorf("fs locker expired unexpected error: %v", err)
		}

		if !exp {
			t.Errorf("expected lock to be expired")
		}
	})
}

func TestRecognizesNonExpiredLock(t *testing.T) {
	execFsTest(t, func(l *FsLocker) {
		key := "test.key"

		_, err := l.Lock(key, 10*time.Second)
		if err != nil {
			t.Errorf("fs locker lock unexpected error: %v", err)
		}

		_, exp, err := l.Expired(key)
		if err != nil {
			t.Errorf("fs locker expired unexpected error: %v", err)
		}

		if exp {
			t.Errorf("expected lock to be non-expired")
		}
	})
}
