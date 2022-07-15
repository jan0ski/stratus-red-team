package main

import (
	"errors"
	"fmt"

	"github.com/datadog/stratus-red-team/pkg/stratus"
	"github.com/spf13/cobra"
)

func buildShowCmd() *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Displays detailed information about an attack technique.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("you must specify at least one attack technique")
			}
			_, err := resolveTechniques(args)
			return err
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return getTechniquesCompletion(toComplete), cobra.ShellCompDirectiveNoFileComp
		},
		Run: func(cmd *cobra.Command, args []string) {
			techniques, _ := resolveTechniques(args)
			doShowCmd(techniques)
		},
	}
	showCmd.AddCommand(buildTerraformCmd())
	return showCmd
}

func buildTerraformCmd() *cobra.Command {
	terraformCmd := &cobra.Command{
		Use:   "terraform",
		Short: "Displays the terraform resources created for an attack technique.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("you must specify at least one attack technique")
			}
			_, err := resolveTechniques(args)
			return err
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return getTechniquesCompletion(toComplete), cobra.ShellCompDirectiveNoFileComp
		},
		Run: func(cmd *cobra.Command, args []string) {
			techniques, _ := resolveTechniques(args)
			doTerraformCmd(techniques)
		},
	}
	return terraformCmd
}

func doTerraformCmd(techniques []*stratus.AttackTechnique) {
	for i := range techniques {
		fmt.Println(string(techniques[i].PrerequisitesTerraformCode))
	}
}

func doShowCmd(techniques []*stratus.AttackTechnique) {
	for i := range techniques {
		fmt.Println(techniques[i].Description)
	}
}
