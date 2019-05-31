// Copyright © 2019 Mikhail Advani mikhail.advani@gmail.com
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

package terramodule

import (
	"archive/zip"
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestSuccessfulArchive(t *testing.T) {
	assert := assert.New(t)
	err := archiveE("testdata/sample_module", "testdata/output/sample_module-successful.zip")
	assert.NoError(err)
}

func TestSuccessfulArchiveChecksum(t *testing.T) {
	assert := assert.New(t)
	ExecPackage("testdata/e2e_test_module", "successful-checksum", "testdata/output")
	checksum := readFileFromArchive("testdata/output/e2e_test_module-successful-checksum.zip", "e2e_test_module/.checksum")
	assert.Equal("e53dd62d38c5e7e2b8a8261c4b4aced1", checksum)
}

func TestUnsuccessfulArchiveForNonExistentDirectory(t *testing.T) {
	assert := assert.New(t)
	err := archiveE("testdata/sample_modul", "testdata/output/sample_module-non-existent-dir.zip")
	assert.EqualError(err, "stat testdata/sample_modul: no such file or directory")
}

func TestUnsuccessfulArchiveForFile(t *testing.T) {
	assert := assert.New(t)
	err := archiveE("testdata/sample_module/main.tf", "testdata/output/sample_module-file.zip")
	assert.EqualError(err, "Input path should be a directory")
}

func TestFileChecksum(t *testing.T) {
	assert := assert.New(t)
	checksum := getFileChecksum("testdata/sample_module/main.tf")
	assert.Equal("f83d72e373c4d172bae35bf03931fc4b", checksum)
}

func TestModuleChecksum(t *testing.T) {
	assert := assert.New(t)
	checksum := getModuleChecksum("testdata/sample_module")
	assert.Equal("e53dd62d38c5e7e2b8a8261c4b4aced1", checksum)
}

func readFileFromArchive(archiveName, filename string) string {
	contents := new(bytes.Buffer)
	r, _ := zip.OpenReader(archiveName)
	defer r.Close()

	for _, f := range r.File {
		if f.Name == filename {
			rc, _ := f.Open()
			io.CopyN(contents, rc, 68)
			rc.Close()
			return contents.String()
		}
	}
	return ""
}
