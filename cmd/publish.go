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
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	publishCmd.Flags().StringP("temp-dir", "", "/tmp", "Temporary directory")
	publishCmd.Flags().StringP("storage", "", "s3", "Object storage")
	publishCmd.Flags().StringP("s3-bucket", "", "", "Target S3 bucket to publish the module to")
	publishCmd.Flags().StringP("s3-prefix", "", "", "Prefix of the module zip")
}

func execPublish(cmd *cobra.Command, args []string) {
	ValidateFlags(cmd.Flags())
	moduleFolder := GetStringFlag(cmd.Flags(), "module")
	tempDir := GetStringFlag(cmd.Flags(), "temp-dir")
	s3Bucket := GetStringFlag(cmd.Flags(), "s3-bucket")
	s3Prefix := GetStringFlag(cmd.Flags(), "s3-prefix")
	zipFileLocation := ZipModule(moduleFolder, tempDir)
	s3Key := path.Join(s3Prefix, path.Base(zipFileLocation))
	PublishIfNotAlreadyOnS3(s3Bucket, s3Key, zipFileLocation)
}

// PublishIfNotAlreadyOnS3 uploads a file to S3 if it is not already present
func PublishIfNotAlreadyOnS3(s3Bucket, s3Key, zipFileLocation string) {
	if !(ObjectOnS3(s3Bucket, s3Key)) {
		fmt.Println("Will publish")
	}
}

// ObjectOnS3 checks if an object exists on S3
func ObjectOnS3(s3Bucket, s3Key string) bool {
	exists, err := ObjectOnS3E(s3Bucket, s3Key)
	if err != nil {
		log.Fatal("Error checking object present on S3")
		log.Fatal(err)
		os.Exit(1)
	}
	return exists
}

// ObjectOnS3E checks if an object exists on S3
func ObjectOnS3E(s3Bucket, s3Key string) (bool, error) {
	svc := s3.New(session.New())
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Key),
	}
	_, err := svc.HeadObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				return false, nil
			default:
				return false, aerr
			}
		} else {
			return false, err
		}
	}
	return true, nil
}

// GetStringFlag gets the value of a string flag with error handling
func GetStringFlag(flags *pflag.FlagSet, flagName string) string {
	value, err := flags.GetString(flagName)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting flag %s", flagName))
		log.Fatal(err)
		os.Exit(1)
	}
	return value
}

// ValidateFlags checks if the flags passed result in a compatible set
func ValidateFlags(flags *pflag.FlagSet) {
	moduleFolder, _ := flags.GetString("module")
	if moduleFolder == "" {
		log.Fatal("Module folder is mandatory")
		os.Exit(1)
	}
	storage, _ := flags.GetString("storage")
	if storage == "s3" {
		s3Bucket, _ := flags.GetString("s3-bucket")
		if s3Bucket == "" {
			log.Fatal("S3 Bucket is mandatory")
			os.Exit(1)
		}
	} else {
		log.Fatal("Only S3 storage is currently supported")
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

// FolderChecksum computes the checksum of a folder(combination of checksums of all files of a folder)
func FolderChecksum(folderPath, tempFile string) string {
	checksum, err := FolderChecksumE(folderPath, tempFile)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return checksum
}

// FolderChecksumE computes the checksum of a folder(combination of checksums of all files of a folder)
func FolderChecksumE(folderPath, tempFile string) (string, error) {
	DeleteFileIfExistsE(tempFile)
	filenames, err := GetFilesOfFolderE(folderPath)
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(tempFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	defer f.Close()
	for _, element := range filenames {
		_, err = f.WriteString(FileChecksum(element) + "\n")
		if err != nil {
			return "", err
		}
	}
	return FileChecksum(tempFile), nil
}

// GetFilesOfFolderE returns all the files in a folder
func GetFilesOfFolderE(folderPath string) ([]string, error) {
	var filenames []string
	err := filepath.Walk(folderPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsRegular() {
				filenames = append(filenames, path)
			}
			return nil
		})
	if err != nil {
		return make([]string, 0), err
	}
	return filenames, nil
}

// DeleteFileIfExistsE will delete a file if it is present
func DeleteFileIfExistsE(filepath string) error {
	if _, err := os.Stat(filepath); err == nil {
		err := os.Remove(filepath)
		if err != nil {
			return err
		}
	}
	return nil
}

// ZipModule creates a Zip of the module with name <FOLDER>-<VERSION>.zip
func ZipModule(sourceFolder, tempDir string) string {
	version := ReadVersion(sourceFolder)
	moduleName := path.Base(sourceFolder)
	zipFileLocation := path.Join(tempDir, fmt.Sprintf("%s-%s.zip", moduleName, version))
	err := DeleteFileIfExistsE(zipFileLocation)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	Zip(sourceFolder, zipFileLocation)
	return zipFileLocation
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
