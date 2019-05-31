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
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ExecPackage runs the package logic
func ExecPackage(moduleDirectory string, moduleVersion string, outputDir string) {
	validate(moduleDirectory)
	moduleName := path.Base(moduleDirectory)
	targetZipFileLocation := fmt.Sprintf("%s/%s-%s.zip", outputDir, moduleName, moduleVersion)
	writeModuleChecksum(moduleDirectory)
	archive(moduleDirectory, targetZipFileLocation)
}

func validate(moduleDirectory string) {
	src, err := os.Stat(moduleDirectory)
	if os.IsNotExist(err) {
		log.Fatal(fmt.Sprintf("%s directory does not exist", moduleDirectory))
	} else if err != nil {
		panic(err)
	}
	if !src.IsDir() {
		log.Fatal(fmt.Sprintf("%s is not a directory", moduleDirectory))
	}
}

func writeModuleChecksum(moduleDirectory string) {
	moduleChecksum := getModuleChecksum(moduleDirectory)
	d1 := []byte(moduleChecksum)
	err := ioutil.WriteFile(fmt.Sprintf("%s/.checksum", moduleDirectory), d1, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func getModuleChecksum(moduleDirectory string) string {
	fileChecksums := ""
	files, err := ioutil.ReadDir(moduleDirectory)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fileChecksums = fileChecksums + getFileChecksum(fmt.Sprintf("%s/%s", moduleDirectory, f.Name()))
	}
	h := md5.New()
	io.WriteString(h, fileChecksums)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func getFileChecksum(filepath string) string {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func archive(sourceFolder, zipFileLocation string) {
	err := archiveE(sourceFolder, zipFileLocation)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func archiveE(sourceFolder, zipFileLocation string) error {
	zipfile, err := os.Create(zipFileLocation)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(sourceFolder)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("Input path should be a directory")
	}

	baseDir := filepath.Base(sourceFolder)

	filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, sourceFolder))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
