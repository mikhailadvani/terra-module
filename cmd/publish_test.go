package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadVersionFile(t *testing.T) {
	t.Parallel()
	actualVersion := ReadVersion("testdata/sample_module")
	expectedVersion := "0.0.1"
	assert.Equal(t, expectedVersion, actualVersion)
}

func TestFileChecksum(t *testing.T) {
	t.Parallel()
	actualChecksum := FileChecksum("testdata/sample_module-0.0.1.zip")
	expectedChecksum := "b75276202a1d854e098a5ee22c7175af"
	assert.Equal(t, expectedChecksum, actualChecksum)
}

func TestFolderChecksum(t *testing.T) {
	t.Parallel()
	actualChecksum := FolderChecksum("testdata/sample_module", "/tmp/test_data.txt")
	expectedChecksum := "8dc6dae213fa6661995dd5811311d2ad"
	assert.Equal(t, expectedChecksum, actualChecksum)
}

func TestZipModule(t *testing.T) {
	t.Parallel()
	moduleDir := "testdata/sample_module"
	tempDir := "/tmp"
	ZipModule(moduleDir, tempDir)
	assert.FileExists(t, "/tmp/sample_module-0.0.1.zip")
}
