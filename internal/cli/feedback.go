package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/feedback"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/hook/security"
)

// feedbackSecurityFileName is the config section the scrubber reads: it carries
// both the user's extra sensitive-content patterns and the extra environment
// variable names whose values are masked.
const feedbackSecurityFileName = "security.yaml"

// newFeedbackCmd constructs the `moai feedback` parent command.
//
// The command tree is a thin wiring layer over internal/feedback: it moves text
// between stdin, stdout and exit codes and owns no transformation of its own.
//
// [HARD] The verdict travels in the stdout JSON; the exit code signals TOOL
// FAILURE ONLY. A policy block is exit 0 with "verdict":"blocked", so the
// caller needs exactly two rules — a non-zero exit means do not submit, and a
// verdict other than "ok" means do not submit — instead of a numeric protocol
// that mixes the two axes.
//
// This CLI surface never prompts the user: it emits JSON on stdout and
// human-readable diagnostics on stderr, and the orchestrator owns every
// interaction.
func newFeedbackCmd() *cobra.Command {
	var root string

	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "Feedback submission helpers (scrub, queue, …)",
		Long: `Feedback submission helpers.

'scrub' masks a feedback report — title and body alike — before it can leave
the machine, and classifies it: a report that reads as a security vulnerability
disclosure is refused for the public issue tracker.

'queue' holds a scrubbed report whose submission failed so it can be re-sent.
The queue's whole read scope is its own queue.json; it never reads the
pre-submit draft, which holds pre-scrub raw text for a different failure.`,
		SilenceUsage: true,
	}
	cmd.PersistentFlags().StringVar(&root, "root", "",
		"project root whose .moai tree receives the mask log and the queue (default: nearest ancestor holding .moai)")

	cmd.AddCommand(newFeedbackScrubCmd(&root))
	cmd.AddCommand(newFeedbackQueueCmd(&root))
	return cmd
}

// newFeedbackScrubCmd wires `moai feedback scrub`: the body arrives on stdin,
// the title on --title, and the whole result leaves as one JSON object.
func newFeedbackScrubCmd(root *string) *cobra.Command {
	var title string

	cmd := &cobra.Command{
		Use:   "scrub",
		Short: "Mask and classify a feedback report read from stdin",
		Long: `Mask and classify a feedback report.

The body is read from stdin and the title from --title. BOTH are scrubbed: the
title is a separate argument of the submission command and is free-form user
text, so a scrubber that only saw the body would leak through it.

stdout carries a single JSON object with verdict, title, body, findings and
reason. Findings report kind, location and count only — never a raw value.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			return runFeedbackScrub(c, *root, title)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "the report title, scrubbed alongside the body")
	return cmd
}

// runFeedbackScrub reads the body, resolves the root and the policy, and emits
// the result.
//
// Every failure before the emit is a tool failure and returns an error: the
// caller must not be able to read "no output" as "nothing to mask".
func runFeedbackScrub(cmd *cobra.Command, rootFlag, title string) error {
	body, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return fmt.Errorf("feedback scrub: reading body from stdin: %w", err)
	}

	root, err := resolveFeedbackRoot(rootFlag, "")
	if err != nil {
		return fmt.Errorf("feedback scrub: %w", err)
	}

	opt, err := feedbackScrubOptions(root)
	if err != nil {
		return fmt.Errorf("feedback scrub: %w", err)
	}

	res, err := feedback.Scrub(feedback.Input{Title: title, Body: string(body)}, opt)
	if err != nil {
		return fmt.Errorf("feedback scrub: %w", err)
	}
	return emitFeedbackJSON(cmd, res)
}

// newFeedbackQueueCmd wires `moai feedback queue` with the minimal verb set the
// skill body calls: enqueue, list, resolve.
func newFeedbackQueueCmd(root *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue",
		Short: "Retry queue for a scrubbed report whose submission failed",
		Long: `Retry queue for a scrubbed report whose submission failed.

[HARD] The queue owns exactly one failure: the submission command returned
non-zero AFTER a submission was attempted, so what it holds is the MASKED text
that was about to be published. A failure BEFORE anything was sent (no
credentials, a rate limit) belongs to the pre-submit draft, which holds
pre-scrub raw text and is never read here.`,
		SilenceUsage: true,
	}

	enqueue := &cobra.Command{
		Use:   "enqueue",
		Short: "Queue a scrubber result read from stdin",
		Long: `Queue a scrubber result read from stdin.

Input is the JSON object 'moai feedback scrub' wrote, verbatim. Taking the
scrubber's own output — rather than a title and a body — is what makes it
impossible to queue raw user text by calling this verb differently. A blocked
result is refused: the queue exists to re-send, and re-sending what the gate
declined would publish it.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			return runFeedbackQueueEnqueue(c, *root)
		},
	}

	list := &cobra.Command{
		Use:          "list",
		Short:        "List the queued reports",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			return runFeedbackQueueList(c, *root)
		},
	}

	resolve := &cobra.Command{
		Use:          "resolve <id>",
		Short:        "Remove a queued report whose re-send succeeded",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runFeedbackQueueResolve(c, *root, args[0])
		},
	}

	cmd.AddCommand(enqueue, list, resolve)
	return cmd
}

func runFeedbackQueueEnqueue(cmd *cobra.Command, rootFlag string) error {
	raw, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return fmt.Errorf("feedback queue enqueue: reading result from stdin: %w", err)
	}
	var res feedback.Result
	if err := json.Unmarshal(raw, &res); err != nil {
		return fmt.Errorf("feedback queue enqueue: parsing scrubber result: %w", err)
	}
	if res.Verdict == "" {
		return fmt.Errorf("feedback queue enqueue: input carries no verdict; expected the JSON written by 'moai feedback scrub'")
	}

	store, err := feedbackQueueStore(rootFlag)
	if err != nil {
		return fmt.Errorf("feedback queue enqueue: %w", err)
	}
	item, err := store.EnqueueMasked(res)
	if err != nil {
		return fmt.Errorf("feedback queue enqueue: %w", err)
	}
	return emitFeedbackJSON(cmd, item)
}

func runFeedbackQueueList(cmd *cobra.Command, rootFlag string) error {
	store, err := feedbackQueueStore(rootFlag)
	if err != nil {
		return fmt.Errorf("feedback queue list: %w", err)
	}
	rec, err := store.Load()
	if err != nil {
		return fmt.Errorf("feedback queue list: %w", err)
	}
	return emitFeedbackJSON(cmd, rec)
}

func runFeedbackQueueResolve(cmd *cobra.Command, rootFlag, id string) error {
	store, err := feedbackQueueStore(rootFlag)
	if err != nil {
		return fmt.Errorf("feedback queue resolve: %w", err)
	}
	removed, err := store.Resolve(id)
	if err != nil {
		return fmt.Errorf("feedback queue resolve: %w", err)
	}
	return emitFeedbackJSON(cmd, map[string]any{"id": id, "removed": removed})
}

// feedbackQueueStore resolves the root and returns the store over its queue
// file. Unlike the scrub path, an unresolved root is an error: the queue verbs
// exist to write, and a write with nowhere to go is a silent loss of the one
// artefact that cannot be regenerated.
func feedbackQueueStore(rootFlag string) (*feedback.QueueStore, error) {
	root, err := resolveFeedbackRoot(rootFlag, "")
	if err != nil {
		return nil, err
	}
	if root == "" {
		return nil, fmt.Errorf("cannot resolve a project root (no %s directory found above the working directory); pass --root", defs.MoAIDir)
	}
	return feedback.NewQueueStore(feedback.QueuePathForRoot(root)), nil
}

// resolveFeedbackRoot resolves the project root the on-disk artefacts are
// written under.
//
// An explicit --root is validated rather than trusted: a mistyped path that
// silently fell back to the walk-up would write another project's .moai tree.
// An absent flag walks up from start (the working directory when empty), and
// finding nothing returns an empty root without an error — the scrub itself
// never depends on a root being known, so a report filed outside any project
// still gets masked.
func resolveFeedbackRoot(rootFlag, start string) (string, error) {
	if rootFlag != "" {
		abs, err := filepath.Abs(rootFlag)
		if err != nil {
			return "", fmt.Errorf("resolving --root %s: %w", rootFlag, err)
		}
		info, err := os.Stat(filepath.Join(abs, defs.MoAIDir))
		if err != nil || !info.IsDir() {
			return "", fmt.Errorf("--root %s is not a project root (no %s directory)", rootFlag, defs.MoAIDir)
		}
		return abs, nil
	}
	if root, ok := feedback.ResolveProjectRoot(start); ok {
		return root, nil
	}
	return "", nil
}

// feedbackSandboxSection is the slice of the security section that names the
// extra environment variables whose values are masked. Only this path is
// declared locally; the pattern extensions are read through the exported
// ExtraSecurityConfig type so the two collections cannot drift.
type feedbackSandboxSection struct {
	Security struct {
		Sandbox struct {
			EnvScrubExtra []string `yaml:"env_scrub_extra"`
		} `yaml:"sandbox"`
	} `yaml:"security"`
}

// feedbackScrubOptions builds the scrub options from the project's security
// section: the default policy extended with the user's extra sensitive-content
// patterns, plus the extra environment names.
//
// [HARD] fail-closed, deliberately unlike the detector's fail-open loader. A
// security section that cannot be read or parsed means the user's extra
// patterns are not applied, so the masking would be weaker than the user
// configured — and the output is about to be published. The detector can
// degrade to defaults because a missed pattern there costs one permitted write;
// here it costs a leak, so the tool fails instead.
func feedbackScrubOptions(root string) (feedback.Options, error) {
	opt := feedback.Options{
		ProjectRoot: root,
		Policy:      hook.DefaultSecurityPolicy(),
	}
	if root == "" {
		return opt, nil
	}

	path := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, feedbackSecurityFileName)
	raw, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			// A project without a security section has no extensions to apply;
			// the built-in policy is the whole policy.
			return opt, nil
		}
		return feedback.Options{}, fmt.Errorf("reading security section %s: %w", path, err)
	}

	var extra security.ExtraSecurityConfig
	if err := yaml.Unmarshal(raw, &extra); err != nil {
		return feedback.Options{}, fmt.Errorf("parsing security section %s: %w", path, err)
	}
	opt.Policy.MergeExtraPatterns(&extra)

	var sandboxSection feedbackSandboxSection
	if err := yaml.Unmarshal(raw, &sandboxSection); err != nil {
		return feedback.Options{}, fmt.Errorf("parsing security section %s: %w", path, err)
	}
	opt.EnvScrubExtra = sandboxSection.Security.Sandbox.EnvScrubExtra

	return opt, nil
}

// emitFeedbackJSON writes one JSON document to stdout. Diagnostics never share
// this stream: the caller reads it with jq, and a stray human-readable line
// would make the parse fail.
func emitFeedbackJSON(cmd *cobra.Command, payload any) error {
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\n", encoded); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}
