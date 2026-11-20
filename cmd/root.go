package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type codedError interface {
	error
	ErrorCode() string
	ExitStatus() int
}

var rootCmd = &cobra.Command{
	Use:   "agmd",
	Short: "A task board for you and your coding agents",
	Long: `agmd - a task board for you and your coding agents, driven from the CLI.

Tasks are markdown files under ~/.agmd/task/<project>/ (override with AGMD_HOME).
The project is derived from the repository you run agmd in; Git worktrees
resolve to their main repository, so parallel agents share one task list.

Get started:
  agmd task new setup-db --content "Set up the database"
  agmd task list
  agmd task web

Run 'agmd task --help' for every task command.`,
	Version: "0.1.0",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if coded, ok := err.(codedError); ok {
			if taskJSON {
				payload := map[string]interface{}{
					"error": map[string]string{
						"code":    coded.ErrorCode(),
						"message": coded.Error(),
					},
				}
				encoded, marshalErr := json.Marshal(payload)
				if marshalErr == nil {
					fmt.Fprintln(os.Stderr, string(encoded))
				} else {
					fmt.Fprintln(os.Stderr, err)
				}
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
			os.Exit(coded.ExitStatus())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// ExecuteArgs runs agmd with the given arguments and returns any error.
// Unlike Execute(), it does not call os.Exit — intended for use in tests.
func ExecuteArgs(args []string) error {
	resetFlags(rootCmd)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

// resetFlags resets all flags on cmd and its subcommands to their default values.
// This is required for in-process testing where the same Cobra command tree is
// reused across multiple Execute() calls and flags would otherwise retain values.
func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Value.Set(f.DefValue) //nolint:errcheck
		f.Changed = false
	})
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		f.Value.Set(f.DefValue) //nolint:errcheck
		f.Changed = false
	})
	for _, sub := range cmd.Commands() {
		resetFlags(sub)
	}
}

func init() {
	// Enable shell completion command (agmd completion bash/zsh/fish/powershell)
}
