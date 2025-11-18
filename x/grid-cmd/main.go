package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Descriptor represents a grid subcommand's execution specification
type Descriptor struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Runtime string `json:"runtime"`
		Entry   string `json:"entry"`
	} `json:"spec"`
}

var rootCmd = &cobra.Command{
	Use:   "grid",
	Short: "Distributed command framework with descriptor-based dispatch",
	Long:  "Grid: Execute commands across decentralized paths using cryptographic descriptors",
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringP("gridpath", "g", "/grid", "Grid path root")
	viper.BindPFlag("gridpath", rootCmd.PersistentFlags().Lookup("gridpath"))

	// Register core commands
	rootCmd.AddCommand(cdCmd())
	rootCmd.AddCommand(printCmd())
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(envCmd())
	rootCmd.AddCommand(posCmd())
}

func initConfig() {
	viper.SetDefault("gridpath", "/grid")
	viper.SetEnvPrefix("GRID")
	viper.AutomaticEnv()
}

func cdCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cd [path]",
		Short: "Navigate grid path",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("path required")
			}
			// XXX This is a placeholder; actual implementation would change the context
			fmt.Printf("Changed to: %s\n", args)
			return nil
		},
	}
}

// XXX This is an example and shouldn't be hardcoded
func printCmd() *cobra.Command {
	var printer string
	cmd := &cobra.Command{
		Use:   "print [file]",
		Short: "Print to grid resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("file required")
			}
			fmt.Printf("Print to %s: %s\n", printer, args)
			return nil
		},
	}
	cmd.Flags().StringVarP(&printer, "printer", "P", "default", "Printer path")
	return cmd
}

// XXX This is an example and shouldn't be hardcoded
func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [service]",
		Short: "Execute grid service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("service required")
			}
			fmt.Printf("Running: %s\n", args)
			return nil
		},
	}
}

func envCmd() *cobra.Command {
	subcmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environment",
	}

	setCmd := &cobra.Command{
		Use:   "set [name] [value]",
		Short: "Set variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("name and value required")
			}
			// XXX this is actually not the right thing to do for grid env vars
			os.Setenv(args[0], args[1])
			fmt.Printf("Set %s=%s\n", args[0], args[1])
			// XXX demo of a subprocess using an environment variable
			democmd := exec.Command("printenv", args[0])
			democmd.Stdout = os.Stdout
			democmd.Stderr = os.Stderr
			democmd.Run()
			return nil
		},
	}
	subcmd.AddCommand(setCmd)
	return subcmd
}

// XXX This is an example and shouldn't be hardcoded
func posCmd() *cobra.Command {
	var motor string
	cmd := &cobra.Command{
		Use:   "pos [position]",
		Short: "Control motor position",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("position required")
			}
			fmt.Printf("Motor %s -> position %s\n", motor, args)
			return nil
		},
	}
	cmd.Flags().StringVarP(&motor, "motor", "m", "", "Motor path")
	return cmd
}

func loadDescriptor(path string) (*Descriptor, error) {
	data, err := os.ReadFile(filepath.Join(path, "descriptor.json"))
	if err != nil {
		return nil, err
	}
	var desc Descriptor
	// XXX this should be cbor not json
	if err := json.Unmarshal(data, &desc); err != nil {
		return nil, err
	}
	return &desc, nil
}

// XXX see x/descriptors/main.go for the correct code
func executeDescriptor(desc *Descriptor, args []string) error {
	if desc.Spec.Runtime != "native" {
		return fmt.Errorf("runtime %s not yet supported", desc.Spec.Runtime)
	}
	cmd := exec.Command(desc.Spec.Entry, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
