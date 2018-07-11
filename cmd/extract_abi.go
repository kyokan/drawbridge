package main

import (
	"os"
	"io/ioutil"
	"encoding/json"
	"path"
	"github.com/spf13/cobra"
	"fmt"
	rawlog "log"
	"errors"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "extract-abi",
		Short: "extracts the ABI from Truffle contracts",
		Run: func(c *cobra.Command, args []string) {
			if err := action(c); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	rootCmd.PersistentFlags().StringSliceP("contracts", "c", nil, "list of contracts to extract from")
	rootCmd.PersistentFlags().StringP("output-dir", "o", "", "directory to output ABI files")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func action(c *cobra.Command) error {
	contracts, err := c.Flags().GetStringSlice("contracts")

	if err != nil {
		return err
	}

	outputDir, err := c.Flags().GetString("output-dir")

	if err != nil {
		return err
	}

	if len(contracts) == 0 {
		return errors.New("a set of input contracts is required")
	}

	if outputDir == "" {
		return errors.New("an output directory is required")
	}

	abis := make(map[string][]byte)

	for _, contract := range contracts {
		contents, err := ioutil.ReadFile(contract)

		if err != nil {
			return err
		}

		data := make(map[string]interface{})

		if err = json.Unmarshal(contents, &data); err != nil {
			return err
		}

		abi, err := json.Marshal(data["abi"])

		if err != nil {
			return err
		}

		abis[path.Base(contract)] = abi
	}

	for k, v := range abis {
		p := path.Join(outputDir, k)
		err := ioutil.WriteFile(p, v, 0744)

		if err != nil {
			return err
		}

		rawlog.Printf("Outputted ABI file %s", k)
	}

	return nil
}