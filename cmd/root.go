package cmd

import (
	"fmt"
	"os"

	"github.com/yourusername/iamctl/cmd/enforce"
	password "github.com/yourusername/iamctl/cmd/password"
	"github.com/yourusername/iamctl/cmd/mfa"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "iamctl",
	Short: "A CLI tool for managing AWS IAM credentials",
	Long: `iamctl is a CLI tool that helps you manage your AWS IAM credentials
including rotating access keys, changing passwords, and managing MFA.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = "0.1.0"
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number")
	
	// Add status command
	rootCmd.AddCommand(NewStatusCommand())
	
	// Add keys commands
	keysCmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage IAM access keys",
	}
	
	keysCmd.AddCommand(NewRotateCommand())
	keysCmd.AddCommand(NewDisableCommand())
	rootCmd.AddCommand(keysCmd)
	
	// Add password commands
	passwordCmd := &cobra.Command{
		Use:   "password",
		Short: "Manage IAM user passwords",
	}
	
	resetCmd := password.NewResetCommand()
	passwordCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(passwordCmd)
	
	// Add MFA commands
	mfaCmd := &cobra.Command{
		Use:   "mfa",
		Short: "Manage MFA devices",
	}
	
	mfaCmd.AddCommand(mfa.NewEnableCommand())
	mfaCmd.AddCommand(mfa.NewDisableCommand())
	mfaCmd.AddCommand(mfa.NewStatusCommand())
	rootCmd.AddCommand(mfaCmd)
	
	// Add enforce commands
	enforceCmd := &cobra.Command{
		Use:   "enforce",
		Short: "Enforce security policies",
	}
	
	enforceCmd.AddCommand(enforce.NewMFACommand())
	enforceCmd.AddCommand(enforce.NewPolicyCommand())
	rootCmd.AddCommand(enforceCmd)
}