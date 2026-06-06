package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/utils"

	"github.com/spf13/cobra"
)

const deprecatedPasswordFlagMessage = "use the interactive password prompt or --password-stdin instead; CLI password flags can leak via shell history and process inspection"

func validatePasswordInputFlags(password string, passwordStdin bool) error {
	if password != "" && passwordStdin {
		return fmt.Errorf("cannot use --password and --password-stdin together")
	}
	return nil
}

func warnDeprecatedPasswordFlag(password string) {
	if password == "" {
		return
	}

	fmt.Fprintf(os.Stderr, "Warning: --password is deprecated and less safe. Prefer the password prompt or --password-stdin.\n")
}

func readPasswordFromStdin(cmd *cobra.Command) (string, error) {
	password, err := utils.ReadSecretFromStdin(cmd.InOrStdin())
	if err != nil {
		return "", fmt.Errorf("read password from stdin: %w", err)
	}
	return password, nil
}
