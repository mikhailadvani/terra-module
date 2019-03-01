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

package cmd

import (
	"archive/zip"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Archive and publish a terraform module to S3",
	Run: func(cmd *cobra.Command, args []string) {
		execPublish(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringP("module", "m", "", "Module folder")
}

func execPublish(cmd *cobra.Command, args []string) {
	moduleFolder, _ := cmd.Flags().GetString("module")
	if moduleFolder == "" {
		fmt.Println("Module folder is mandatory")
		os.Exit(1)
	}
}

// ReadVersion reads the VERSION.txt file from the module folder
func ReadVersion(moduleFolder string) string {
	content, err := ReadVersionE(moduleFolder)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return content
}

// ReadVersionE reads the VERSION.txt file from the module folder
func ReadVersionE(moduleFolder string) (string, error) {
	versionFile := fmt.Sprintf("%s/VERSION.txt", moduleFolder)
	content, err := ioutil.ReadFile(versionFile)
	if err != nil {
		return "", err
	}
	stringifiedContent := string(content)
	strippedContent := strings.TrimSuffix(stringifiedContent, "\n")
	return strippedContent, nil
}

// FileChecksum computes the checksum of a file
func FileChecksum(filename string) string {
	checksum, err := FileChecksumE(filename)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return checksum
}

// FileChecksumE computes the checksum of a file
func FileChecksumE(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// Zip packages a folder to a zip file
func Zip(sourceFolder, zipFileLocation string) {
	err := ZipE(sourceFolder, zipFileLocation)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

// ZipE packages a folder to a zip file
func ZipE(sourceFolder, zipFileLocation string) error {
	// zipFileLocation := fmt.Sprintf("%s.zip", sourceFolder)
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
