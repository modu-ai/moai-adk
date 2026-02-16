// Package github provides GitHub CLI integration for PR and issue operations.
//
// It wraps the gh CLI binary (https://cli.github.com/) via os/exec to provide
// pull request creation, review, merge, and CI/CD status checking. All GitHub
// interactions are abstracted behind testable interfaces.
package github
