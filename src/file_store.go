// Copyright 2018 Bull S.A.S. Atos Technologies - Bull, Rue Jean Jaures, B.P.68, 78340, Les Clayes-sous-Bois, France.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ystia/yorc/v4/storage/encoding"
	"github.com/ystia/yorc/v4/storage/types"
	"golang.org/x/sync/errgroup"
)

// Options are the options for the Go map store.
type options struct {
	// The directory in which to store files.
	// Can be absolute or relative.
	directory string
	// Extension of the filename, e.g. "json".
	filenameExtension string
	// Encoding format.
	// Note: When you change this, you should also change the FilenameExtension if it's not empty ("").
	codec encoding.Codec
}

type fileStore struct {
	// For locking the locks map
	// (no two goroutines may create a lock for a filename that doesn't have a lock yet).
	locksLock *sync.Mutex
	// For locking file access.
	fileLocks         map[string]*sync.RWMutex
	filenameExtension string
	directory         string
	codec             encoding.Codec
}

// prepareFileLock returns an existing file lock or creates a new one
func (s *fileStore) prepareFileLock(filePath string) *sync.RWMutex {
	s.locksLock.Lock()
	lock, found := s.fileLocks[filePath]
	if !found {
		lock = new(sync.RWMutex)
		s.fileLocks[filePath] = lock
	}
	s.locksLock.Unlock()
	return lock
}

func (s *fileStore) buildFilePath(k string, withExtension bool) string {
	filePath := k
	if withExtension && s.filenameExtension != "" {
		filePath += "." + s.filenameExtension
	}
	return filepath.Join(s.directory, filePath)
}

func (s *fileStore) Set(ctx context.Context, k string, v interface{}) error {
	if err := types.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	data, err := s.codec.Marshal(v)
	if err != nil {
		return err
	}

	// Prepare file lock.
	filePath := s.buildFilePath(k, true)
	lock := s.prepareFileLock(filePath)

	// File lock and file handling.
	lock.Lock()
	defer lock.Unlock()

	err = os.MkdirAll(path.Dir(filePath), 0700)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, data, 0600)
}

func (s *fileStore) SetCollection(ctx context.Context, keyValues []*types.KeyValue) error {
	if keyValues == nil {
		return nil
	}
	errGroup, ctx := errgroup.WithContext(ctx)
	for _, kv := range keyValues {
		kvItem := kv
		errGroup.Go(func() error {
			return s.Set(ctx, kvItem.Key, kvItem.Value)
		})
	}

	return errGroup.Wait()
}

func (s *fileStore) Get(k string, v interface{}) (bool, error) {
	if err := types.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	filePath := s.buildFilePath(k, true)

	// Prepare file lock.
	lock := s.prepareFileLock(filePath)

	// File lock and file handling.
	lock.RLock()
	// Deferring the unlocking would lead to the unmarshalling being done during the lock, which is bad for performance.
	data, err := ioutil.ReadFile(filePath)
	lock.RUnlock()
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, s.codec.Unmarshal(data, v)
}

func (s *fileStore) Exist(k string) (bool, error) {
	filePath := s.buildFilePath(k, true)

	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *fileStore) Keys(k string) ([]string, error) {
	log.Printf("[FileStore Plugin] Retrieving sub-keys from key:%q", k)
	filePath := s.buildFilePath(k, false)

	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	result := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			result = append(result, strings.TrimSuffix(fileName, path.Ext(fileName)))
		}
		file.Name()
	}

	return result, nil
}

func (s *fileStore) Delete(ctx context.Context, k string, recursive bool) error {
	log.Printf("[FileStore Plugin] Deleting key:%q, recursive:%t", k, recursive)
	if err := types.CheckKey(k); err != nil {
		return err
	}

	filePath := s.buildFilePath(k, !recursive)

	// Prepare file lock.
	lock := s.prepareFileLock(filePath)

	// File lock and file handling.
	lock.Lock()
	defer lock.Unlock()

	var err error
	if !recursive {
		err = os.Remove(filePath)
	} else {
		err = os.RemoveAll(filePath)
	}

	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *fileStore) Types() []types.StoreType {
	t := make([]types.StoreType, 0)
	t = append(t, types.StoreTypeDeployment)
	return t
}
