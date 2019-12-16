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
	"fmt"
	"github.com/ystia/yorc/v4/commands"
	"github.com/ystia/yorc/v4/storage/encoding"
	"log"
	"os"
	"path"
	"sync"
)

// exported as symbol plugin named "Store"
var Store fileStore

func init() {
	var err error

	configuration := commands.GetConfig()
	options := options{
		directory:         path.Join(configuration.WorkingDirectory, "store"),
		codec:             encoding.JSON,
		filenameExtension: "json",
	}
	Store, err = NewFileStore(options)
	if err != nil {
		fmt.Print("failed to create new file store client")
	}
}

// NewFileStore creates a new file store
func NewFileStore(options options) (fileStore, error) {
	log.Printf("*[FileStore Plugin] Create new File store")
	result := fileStore{}

	err := os.MkdirAll(options.directory, 0700)
	if err != nil {
		return result, err
	}

	result.directory = options.directory
	result.locksLock = new(sync.Mutex)
	result.fileLocks = make(map[string]*sync.RWMutex)
	result.filenameExtension = options.filenameExtension
	result.codec = options.codec

	return result, nil
}
