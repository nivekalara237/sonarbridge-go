package cli

import "github.com/spf13/cobra"

type CLIError struct {
	Err       error
	ShowUsage bool
	Command   *cobra.Command
}

func (e *CLIError) Error() string {
	return e.Err.Error()
}

func (e *CLIError) Unwrap() error {
	return e.Err
}
