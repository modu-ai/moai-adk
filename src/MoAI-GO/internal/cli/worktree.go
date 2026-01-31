package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/anthropics/moai-adk-go/internal/output"
	"github.com/anthropics/moai-adk-go/internal/worktree"
	"github.com/spf13/cobra"
)

// NewWorktreeCommand creates the worktree command group
func NewWorktreeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "worktree",
		Aliases: []string{"wt"},
		Short:   "Manage Git worktrees for parallel SPEC development",
		Long:    `Create, manage, and switch between Git worktrees for isolated parallel development of SPEC documents.`,
	}

	cmd.AddCommand(
		newWorktreeNewCmd(),
		newWorktreeListCmd(),
		newWorktreeGoCmd(),
		newWorktreeDoneCmd(),
		newWorktreeRemoveCmd(),
		newWorktreeSyncCmd(),
		newWorktreeCleanCmd(),
		newWorktreeRecoverCmd(),
		newWorktreeStatusCmd(),
	)

	return cmd
}

func getManager() (*worktree.Manager, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	return worktree.NewManager(cwd)
}

func newWorktreeNewCmd() *cobra.Command {
	var (
		branch     string
		baseBranch string
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "new SPEC_ID",
		Short: "Create a new worktree for a SPEC",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := getManager()
			if err != nil {
				return err
			}

			info, err := mgr.Create(args[0], branch, baseBranch, force)
			if err != nil {
				return err
			}

			fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Worktree created: %s", info.SpecID)))
			fmt.Println(output.MutedStyle.Render(fmt.Sprintf("  Path:   %s", info.Path)))
			fmt.Println(output.MutedStyle.Render(fmt.Sprintf("  Branch: %s", info.Branch)))
			return nil
		},
	}

	cmd.Flags().StringVarP(&branch, "branch", "b", "", "custom branch name (default: feature/<SPEC_ID>)")
	cmd.Flags().StringVar(&baseBranch, "base", "main", "base branch to create from")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force creation if exists")
	return cmd
}

func newWorktreeListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active worktrees",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := getManager()
			if err != nil {
				return err
			}

			items := mgr.List()
			if len(items) == 0 {
				fmt.Println(output.MutedStyle.Render("No worktrees found."))
				return nil
			}

			if jsonOutput {
				data, marshalErr := json.MarshalIndent(items, "", "  ")
				if marshalErr != nil {
					return marshalErr
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println(output.HeaderStyle.Render("Active Worktrees"))
			fmt.Println()
			for _, item := range items {
				status := output.SuccessStyle.Render("●")
				if item.Status == "recovered" {
					status = output.WarningStyle.Render("●")
				}
				fmt.Printf("  %s %s\n", status, output.InfoStyle.Render(item.SpecID))
				fmt.Printf("    Branch: %s\n", item.Branch)
				fmt.Printf("    Path:   %s\n", item.Path)
				if item.CreatedAt != "" {
					fmt.Printf("    Created: %s\n", item.CreatedAt)
				}
				fmt.Println()
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
	return cmd
}

func newWorktreeGoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "go SPEC_ID",
		Short: "Open shell in a worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := getManager()
			if err != nil {
				return err
			}

			info := mgr.Registry.Get(args[0])
			if info == nil {
				fmt.Println(output.ErrorStyle.Render(fmt.Sprintf("Worktree not found: %s", args[0])))
				items := mgr.List()
				if len(items) > 0 {
					fmt.Println(output.MutedStyle.Render("  Available worktrees:"))
					for _, item := range items {
						fmt.Printf("    - %s (%s)\n", item.SpecID, item.Branch)
					}
				}
				return fmt.Errorf("worktree not found: %s", args[0])
			}

			if _, statErr := os.Stat(info.Path); os.IsNotExist(statErr) {
				return fmt.Errorf("worktree path does not exist: %s", info.Path)
			}

			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/bash"
			}

			fmt.Println(output.InfoStyle.Render(fmt.Sprintf("Opening shell in %s ...", info.Path)))
			fmt.Println(output.MutedStyle.Render("  Type 'exit' to return."))

			proc := os.ProcAttr{
				Dir:   info.Path,
				Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
			}
			p, procErr := os.StartProcess(shell, []string{shell}, &proc)
			if procErr != nil {
				return fmt.Errorf("failed to start shell: %w", procErr)
			}
			_, waitErr := p.Wait()
			return waitErr
		},
	}
}

func newWorktreeDoneCmd() *cobra.Command {
	var (
		baseBranch string
		push       bool
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "done SPEC_ID",
		Short: "Merge worktree branch and clean up",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := getManager()
			if err != nil {
				return err
			}

			result, doneErr := mgr.Done(args[0], baseBranch, push, force)
			if doneErr != nil {
				return doneErr
			}

			fmt.Println(output.SuccessStyle.Render("✓ Worktree completed"))
			fmt.Println(output.MutedStyle.Render(fmt.Sprintf("  Merged: %s → %s", result["merged_branch"], result["base_branch"])))
			if result["pushed"] == "true" {
				fmt.Println(output.MutedStyle.Render("  Pushed to remote"))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&baseBranch, "base", "main", "base branch to merge into")
	cmd.Flags().BoolVar(&push, "push", false, "push to remote after merge")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force removal with uncommitted changes")
	return cmd
}

func newWorktreeRemoveCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove SPEC_ID",
		Short: "Remove a worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := getManager()
			if err != nil {
				return err
			}

			if removeErr := mgr.Remove(args[0], force); removeErr != nil {
				return removeErr
			}

			fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Worktree removed: %s", args[0])))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force removal")
	return cmd
}

func newWorktreeSyncCmd() *cobra.Command {
	var (
		baseBranch string
		rebase     bool
	)

	cmd := &cobra.Command{
		Use:   "sync SPEC_ID",
		Short: "Sync worktree with base branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := getManager()
			if err != nil {
				return err
			}

			if syncErr := mgr.Sync(args[0], baseBranch, rebase); syncErr != nil {
				return syncErr
			}

			fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Synced %s with %s", args[0], baseBranch)))
			return nil
		},
	}

	cmd.Flags().StringVar(&baseBranch, "base", "main", "base branch to sync from")
	cmd.Flags().BoolVar(&rebase, "rebase", false, "use rebase instead of merge")
	return cmd
}

func newWorktreeCleanCmd() *cobra.Command {
	var mergedOnly bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove merged worktrees",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := getManager()
			if err != nil {
				return err
			}

			if mergedOnly {
				count, cleanErr := mgr.CleanMerged("main")
				if cleanErr != nil {
					return cleanErr
				}
				fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Cleaned %d merged worktrees", count)))
				return nil
			}

			// Remove all
			items := mgr.List()
			if len(items) == 0 {
				fmt.Println(output.MutedStyle.Render("No worktrees to clean."))
				return nil
			}
			count := 0
			for _, item := range items {
				if removeErr := mgr.Remove(item.SpecID, true); removeErr == nil {
					count++
				}
			}
			fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Cleaned %d worktrees", count)))
			return nil
		},
	}

	cmd.Flags().BoolVar(&mergedOnly, "merged-only", false, "only remove worktrees for merged branches")
	return cmd
}

func newWorktreeRecoverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "recover",
		Short: "Recover worktree registry from disk",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := getManager()
			if err != nil {
				return err
			}

			count, recoverErr := mgr.Recover()
			if recoverErr != nil {
				return recoverErr
			}

			if count == 0 {
				fmt.Println(output.MutedStyle.Render("No worktrees to recover."))
			} else {
				fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Recovered %d worktrees", count)))
			}
			return nil
		},
	}
}

func newWorktreeStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show worktree status",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := getManager()
			if err != nil {
				return err
			}

			items := mgr.List()
			if len(items) == 0 {
				fmt.Println(output.MutedStyle.Render("No active worktrees."))
				return nil
			}

			fmt.Println(output.HeaderStyle.Render("Worktree Status"))
			fmt.Println()
			for _, item := range items {
				exists := "exists"
				if _, statErr := os.Stat(item.Path); os.IsNotExist(statErr) {
					exists = "missing"
				}
				fmt.Printf("  %s  branch:%s  path:%s  [%s]\n",
					output.InfoStyle.Render(item.SpecID),
					item.Branch,
					item.Path,
					exists,
				)
			}
			return nil
		},
	}
}
