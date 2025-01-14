/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabtoken

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/hyperledger-labs/fabric-token-sdk/token/core/cmd/pp/cc"
	"github.com/hyperledger-labs/fabric-token-sdk/token/core/cmd/pp/common"
	"github.com/hyperledger-labs/fabric-token-sdk/token/core/fabtoken"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	// Driver is the Token-SDK driver to use
	Driver string
	// OutputDir is the directory to output the generated files
	OutputDir string
	// GenerateCCPackage is whether to generate the chaincode package
	GenerateCCPackage bool
	// Issuers is the list of issuers to include in the public parameters.
	// Each issuer should be specified in the form of <MSP-Dir>:<MSP-ID>
	Issuers []string
	// Auditors is the list of auditors to include in the public parameters.
	// Each auditor should be specified in the form of <MSP-Dir>:<MSP-ID>
	Auditors []string
)

// Cmd returns the Cobra Command for Version
func Cmd() *cobra.Command {
	// Set the flags on the node start command.
	flags := cobraCommand.Flags()
	flags.StringVarP(&OutputDir, "output", "o", ".", "output folder")
	flags.BoolVarP(&GenerateCCPackage, "cc", "", false, "generate chaincode package")
	flags.StringSliceVarP(&Auditors, "auditors", "a", nil, "list of auditor keys in the form of <MSP-Dir>:<MSP-ID>")
	flags.StringSliceVarP(&Issuers, "issuers", "s", nil, "list of issuer keys in the form of <MSP-Dir>:<MSP-ID>")
	return cobraCommand
}

var cobraCommand = &cobra.Command{
	Use:   "fabtoken",
	Short: "Gen FabToken public parameters.",
	Long:  `Generates FabToken public parameters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("trailing args detected")
		}
		// Parsing of the command line is done so silence cmd usage
		cmd.SilenceUsage = true
		raw, err := Gen(&GeneratorArgs{
			OutputDir:         OutputDir,
			GenerateCCPackage: GenerateCCPackage,
			Issuers:           Issuers,
			Auditors:          Auditors,
		})
		if err != nil {
			return errors.Wrap(err, "failed to generate public parameters")
		}
		// generate the chaincode package
		if GenerateCCPackage {
			fmt.Println("Generate chaincode package...")
			if err := cc.GeneratePackage(raw, OutputDir); err != nil {
				return err
			}
		}
		return nil
	},
}

type GeneratorArgs struct {
	// OutputDir is the directory to output the generated files
	OutputDir string
	// GenerateCCPackage is whether to generate the chaincode package
	GenerateCCPackage bool
	// Issuers is the list of issuers to include in the public parameters.
	// Each issuer should be specified in the form of <MSP-Dir>:<MSP-ID>
	Issuers []string
	// Auditors is the list of auditors to include in the public parameters.
	// Each auditor should be specified in the form of <MSP-Dir>:<MSP-ID>
	Auditors []string
}

// Gen generates the public parameters for the FabToken driver
func Gen(args *GeneratorArgs) ([]byte, error) {
	// Setup
	pp, err := fabtoken.Setup()
	if err != nil {
		return nil, errors.Wrap(err, "failed setting up public parameters")
	}
	if err := common.SetupIssuersAndAuditors(pp, args.Auditors, args.Issuers); err != nil {
		return nil, err
	}
	// Store Public Params
	raw, err := pp.Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "failed serializing public parameters")
	}
	path := filepath.Join(args.OutputDir, "fabtoken_pp.json")
	if err := ioutil.WriteFile(path, raw, 0755); err != nil {
		return nil, errors.Wrap(err, "failed writing public parameters to file")
	}

	return raw, nil
}
