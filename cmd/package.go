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
	"github.com/mikhailadvani/terra-module/pkg/terramodule"
	"github.com/spf13/cobra"
)

var moduleVersion string
var outputDir string

// packageCmd represents the package command
var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Archive a terraform module",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		terramodule.ExecPackage(args[0], moduleVersion, outputDir)
	},
}

func init() {
	rootCmd.AddCommand(packageCmd)
	packageCmd.Flags().StringVarP(&moduleVersion, "version", "", "", "Version of the module.")
	packageCmd.MarkFlagRequired("version")
	packageCmd.Flags().StringVarP(&outputDir, "output-dir", "o", ".", "Output directory for the packaged module")
}
