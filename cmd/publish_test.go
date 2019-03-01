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
	actualChecksum := FileChecksum("testdata/sample_module.zip")
	expectedChecksum := "b75276202a1d854e098a5ee22c7175af"
	assert.Equal(t, expectedChecksum, actualChecksum)
}

func TestZip(t *testing.T) {
	t.Parallel()
	sourceFolder := "testdata/sample_module"
	zipFileLocation := "testdata/output/sample_module_2.zip"
	Zip(sourceFolder, zipFileLocation)
	actualChecksum := FileChecksum(zipFileLocation)
	expectedChecksum := "b75276202a1d854e098a5ee22c7175af"
	assert.Equal(t, expectedChecksum, actualChecksum)
}
