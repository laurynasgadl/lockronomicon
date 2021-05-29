package locker

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const metadataFilename = "metadata"

type FsLocker struct {
	rootDir string
}

func NewFsLocker(rootDir string) (*FsLocker, error) {
	err := os.MkdirAll(rootDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &FsLocker{
		rootDir: rootDir,
	}, nil
}

func (fs *FsLocker) Lock(key string, ttl time.Duration) (int64, error) {
	path := filepath.Join(fs.rootDir, key)

	// acquire lock
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		return 0, ErrLockTaken
	}

	// prepare metadata info string
	metadata, err := NewMetadata(ttl).Encode()
	if err != nil {
		// couldn't encode metadata - remove lock
		os.RemoveAll(path)
		return 0, ErrEncodeMetadata
	}

	// write metadata info string to file
	err = os.WriteFile(filepath.Join(path, metadataFilename), metadata, os.ModePerm)
	if err != nil {
		// couldn't write metadata file - remove lock
		os.RemoveAll(path)
		return 0, ErrWriteMetadata
	}

	// get lock dir stats for generation number
	dir, err := os.Stat(path)
	if err != nil {
		// couldn't get lock dir stats - remove lock
		os.RemoveAll(path)
		return 0, ErrReadLock
	}

	return fs.getGenerationNumber(dir), nil
}

func (fs *FsLocker) Refresh(key string, generation int64) (int64, error) {
	path := filepath.Join(fs.rootDir, key)

	// get lock dir stats for generation number comparison
	dir, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, ErrLockNotExist
		} else {
			return 0, ErrReadLock
		}
	}

	if generation != fs.getGenerationNumber(dir) {
		return 0, ErrGenNumberMismatch
	}

	// read current metadata and update it
	metadataPath := filepath.Join(path, metadataFilename)
	mdinfo, err := os.ReadFile(metadataPath)
	if err != nil {
		return 0, ErrReadMetadata
	}

	metadata, err := ParseMetadata(mdinfo)
	if err != nil {
		return 0, ErrDecodeMetadata
	}

	newMetadata, err := NewMetadata(time.Duration(metadata.TTL) * time.Second).Encode()
	if err != nil {
		return 0, ErrEncodeMetadata
	}

	// remove old metadata file and create a new one to force dir's Mod timestamp update
	err = os.Remove(metadataPath)
	if err != nil {
		return 0, ErrRemoveMetadata
	}

	err = os.WriteFile(metadataPath, newMetadata, os.ModePerm)
	if err != nil {
		os.RemoveAll(path)
		return 0, ErrWriteMetadata
	}

	dir, err = os.Stat(path)
	if err != nil {
		return 0, ErrReadLock
	}

	return fs.getGenerationNumber(dir), nil
}

func (fs *FsLocker) Release(key string, generation int64) error {
	path := filepath.Join(fs.rootDir, key)

	dir, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrLockNotExist
		} else {
			return ErrReadLock
		}
	}

	if generation != fs.getGenerationNumber(dir) {
		return ErrGenNumberMismatch
	}

	err = os.RemoveAll(path)
	if err != nil {
		return ErrRemoveLock
	}

	return nil
}

func (fs *FsLocker) Expired(key string) (int64, bool, error) {
	path := filepath.Join(fs.rootDir, key)

	dir, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, ErrLockNotExist
		} else {
			return 0, false, ErrReadLock
		}
	}

	gen := fs.getGenerationNumber(dir)

	metadataPath := filepath.Join(path, metadataFilename)
	mdinfo, err := os.ReadFile(metadataPath)
	if err != nil {
		return 0, false, ErrReadMetadata
	}

	metadata, err := ParseMetadata(mdinfo)
	if err != nil {
		return 0, false, ErrDecodeMetadata
	}

	expired := metadata.Expires != -1 && metadata.Expires <= time.Now().Unix()

	return gen, expired, nil
}

func (fs *FsLocker) getGenerationNumber(file fs.FileInfo) int64 {
	return file.ModTime().UnixNano()
}

type Metadata struct {
	TTL     int64 `json:"ttl"`
	Expires int64 `json:"expires"`
}

func ParseMetadata(data []byte) (*Metadata, error) {
	var md Metadata
	err := json.Unmarshal(data, &md)
	if err != nil {
		return nil, err
	}
	return &md, nil
}

func NewMetadata(ttl time.Duration) *Metadata {
	var expires int64

	if ttl < -1 {
		ttl = -1 * time.Second
	}

	if ttl < 0 {
		expires = -1
	} else {
		expires = time.Now().Add(ttl).Unix()
	}

	return &Metadata{
		TTL:     int64(ttl.Seconds()),
		Expires: expires,
	}
}

func (md *Metadata) Encode() ([]byte, error) {
	return json.Marshal(md)
}
