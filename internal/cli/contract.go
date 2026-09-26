package cli

// contract.go — `moai contract sign|show|verify` (SPEC-AUTONOMY-CONTRACT-001
// M6). RED stub: the command tree is registered so the CLI tests compile;
// every subcommand reports "not implemented".

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// contractGetenvFn is the environment the contract command reads (agent
// markers, the registry override). Tests replace it so the runtime's own
// environment never leaks into a fixture run.
var contractGetenvFn = os.Getenv

// newContractLineReader returns the confirmation reader over the command's
// input stream: one line per call, without the terminator.
var newContractLineReader = func(r io.Reader) func() (string, error) {
	br := bufio.NewReader(r)
	return func() (string, error) {
		line, err := br.ReadString('\n')
		if err != nil && (!errors.Is(err, io.EOF) || line == "") {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}
}

var errContractNotImplemented = errors.New("moai contract: not implemented")

func newContractCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "contract", Short: "Sign, show, and verify SPEC autonomy contracts"}
	var signer, receipt string
	var resign, showJSON, verifyJSON bool
	signCmd := &cobra.Command{Use: "sign <SPEC-ID>...", RunE: func(*cobra.Command, []string) error { return errContractNotImplemented }}
	signCmd.Flags().StringVar(&signer, "signer", "", "")
	signCmd.Flags().StringVar(&receipt, "receipt", "", "")
	signCmd.Flags().BoolVar(&resign, "resign", false, "")
	showCmd := &cobra.Command{Use: "show <SPEC-ID>", RunE: func(*cobra.Command, []string) error { return errContractNotImplemented }}
	showCmd.Flags().BoolVar(&showJSON, "json", false, "")
	verifyCmd := &cobra.Command{Use: "verify <SPEC-ID>", RunE: func(*cobra.Command, []string) error { return errContractNotImplemented }}
	verifyCmd.Flags().BoolVar(&verifyJSON, "json", false, "")
	cmd.AddCommand(signCmd, showCmd, verifyCmd)
	return cmd
}

func init() {
	rootCmd.AddCommand(newContractCmd())
}
