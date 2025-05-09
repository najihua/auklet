// Copyright (c) 2015 Rackspace
// Copyright (c) 2016-2018 iQIYI.com.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// +build !linux

package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

var FileMaxUnsynedBytes = 4 * 1024 * 1024

// TempFile implements an atomic file write by writing to a temp
// directory and then renaming into place.
type TempFile struct {
	*os.File
	saved        bool
	unSyncedSize int
}

// Write temp file with sync mode
func (o *TempFile) Write(b []byte) (int, error) {
	nw, err := o.File.Write(b)
	o.unSyncedSize += nw
	if o.unSyncedSize >= FileMaxUnsynedBytes {
		o.File.Sync()
		o.unSyncedSize = 0
	}
	return nw, err
}

// Abandon removes any resources associated with this file,
// if it hasn't already been saved.
func (o *TempFile) Abandon() error {
	if o.saved {
		return nil
	}
	os.Remove(o.Name())
	return o.File.Close()
}

// Save atomically writes the file to its destination.
func (o *TempFile) Save(dst string) error {
	defer o.File.Close()
	if err := o.File.Sync(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	if err := os.Rename(o.File.Name(), dst); err != nil {
		return err
	}
	o.saved = true
	return nil
}

// Preallocate pre-allocates space for the file.
func (o *TempFile) Preallocate(size int64, reserve int64) error {
	// TODO: this could be done for most non-linux operating systems,
	// but it hasn't been important.
	return nil
}

// NewAtomicFileWriter returns an AtomicFileWriter,
// which handles atomically writing files.
func NewAtomicFileWriter(
	tempDir string, dstDir string) (AtomicFileWriter, error) {
	if err := os.MkdirAll(tempDir, 0770); err != nil {
		return nil, err
	}
	tempFile, err := ioutil.TempFile(tempDir, "")
	if err != nil {
		return nil, err
	}
	return &TempFile{File: tempFile, saved: false, unSyncedSize: 0}, nil
}
