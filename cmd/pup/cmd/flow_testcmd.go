/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"

	"github.com/pupload/pupload/internal/cli/project"
	"github.com/pupload/pupload/internal/cli/run"
	"github.com/pupload/pupload/internal/cli/ui"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test <flowname>",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flow_name := args[0]

		root, err := project.GetProjectRoot()
		if err != nil {
			return err
		}

		go func() {
			log.SetOutput(io.Discard)
			run.RunDevSilent(root)
		}()

		run, err := project.TestFlow(root, "http://localhost:1234/", flow_name)
		if err != nil {
			return err
		}
		fmt.Printf("shit failed")

		ui.TestFlowUI(*run)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	testCmd.PersistentFlags()

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
