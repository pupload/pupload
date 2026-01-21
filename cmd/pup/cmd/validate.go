package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/pupload/pupload/internal/cli/project"
	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/validation"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [flow]",
	Short: "Validate flows in the project",
	Long: `Validate flows in the project against their node definitions.

If a flow name is provided, only that flow will be validated.
Otherwise, all flows in the project will be validated.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := project.GetProjectRoot()
		if err != nil {
			return fmt.Errorf("not inside a project")
		}

		nodeDefs, err := project.GetNodeDefs(root)
		if err != nil {
			return fmt.Errorf("failed to load node definitions: %w", err)
		}

		var flows []models.Flow

		if len(args) > 0 {
			// Validate specific flow
			flow, err := project.GetFlow(root, args[0])
			if err != nil {
				return err
			}
			flows = append(flows, *flow)
		} else {
			// Validate all flows
			flows, err = project.GetFlows(root)
			if err != nil {
				return fmt.Errorf("failed to load flows: %w", err)
			}

			if len(flows) == 0 {
				fmt.Println("No flows found in project")
				return nil
			}
		}

		hasErrors := false

		for _, flow := range flows {
			result := validation.Validate(flow, nodeDefs)
			printValidationResult(flow.Name, result)

			if result.HasError() {
				hasErrors = true
			}
		}

		if hasErrors {
			os.Exit(1)
		}

		return nil
	},
}

var (
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	grey   = color.New(color.FgHiBlack)
	bold   = color.New(color.Bold)
)

func printValidationResult(flowName string, result *validation.ValidationResult) {
	if !result.HasError() && !result.HasWarnings() {
		green.Print("✓ ")
		bold.Print(flowName)
		fmt.Println(": valid")
		return
	}

	red.Print("✗ ")
	bold.Print(flowName)
	fmt.Println(":")

	for _, err := range result.Errors {
		red.Print("  ERROR ")
		grey.Printf("[%s] ", err.Code)
		fmt.Printf("%s: %s\n", err.Name, err.Description)
	}

	for _, warn := range result.Warnings {
		yellow.Print("  WARN  ")
		grey.Printf("[%s] ", warn.Code)
		fmt.Printf("%s: %s\n", warn.Name, warn.Description)
	}

	fmt.Println()
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
