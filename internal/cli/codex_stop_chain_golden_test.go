package cli

// SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2d — AC-HPR-002 (REQ-HPR-002, REQ-HPR-005,
// REQ-HPR-019 cap). Each golden runs the same input through the Claude path and
// the Codex path, with the same project configuration and environment, and
// compares the normalized decision and the reason class under the design
// §D3.4 mapping. The Claude path is the real Claude member: the distributed
// sync-gate script run by bash, HandleCodexReviewGate with a stub reviewer
// session, HandleMultiReviewGate, and the stop-goal evaluator with the real
// command runner. The Codex path is the Codex Stop chain's member.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// wantPair asserts one golden: the normalized decisions are equal, and the
// Codex reason class is the Claude class or its declared §D3.4 pair.
func wantPair(t *testing.T, claude codexadapter.Decision, codex stopMemberOutcome, want codexadapter.Decision, wantClass string) {
	t.Helper()
	if claude != want {
		t.Fatalf("Claude path decision = %s, want %s (the golden's premise does not hold)", claude, want)
	}
	if codex.Decision != claude {
		t.Fatalf("Codex path decision = %s (class %q, reason %q), Claude path = %s", codex.Decision, codex.Class, codex.Reason, claude)
	}
	if codex.Class != wantClass {
		t.Fatalf("Codex reason class = %q, want %q (reason %q)", codex.Class, wantClass, codex.Reason)
	}
}

func TestStopChainEffectParityGolden(t *testing.T) {
	ctx := context.Background()
	fakeCodexVersion(t, "codex-cli 0.0.0-golden")

	// ── goal (member 3) ──────────────────────────────────────────────────
	t.Run("goal", func(t *testing.T) {
		cases := []struct {
			name      string
			cmd       string
			receipt   *int // recorded exit, nil = no receipt
			noGoal    bool
			want      codexadapter.Decision
			wantClass string
		}{
			{name: "unmet", cmd: "false", receipt: intp(1), want: codexadapter.DecisionDeny, wantClass: reasonUnmet},
			{name: "met", cmd: "true", receipt: intp(0), want: codexadapter.DecisionAllow},
			{name: "receipt absent", cmd: "false", want: codexadapter.DecisionDeny, wantClass: reasonUnmeasured},
			{name: "no goal", noGoal: true, want: codexadapter.DecisionAllow},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				f := newStopFixture(t)
				arm := func(session string) {
					if tc.noGoal {
						return
					}
					g := &goal.Goal{SessionID: session, Goal: "golden", Status: goal.StatusArmed,
						Conditions: []goal.Condition{{Type: goal.ConditionMechanical, Cmd: tc.cmd}},
						Ceiling:    goal.Ceiling{MaxTurns: 30}, ProgressionMode: goal.ProgressionAutonomous,
						CreatedAt: time.Now().UTC().Format(time.RFC3339)}
					if err := goal.SaveGoal(f.root, g); err != nil {
						t.Fatal(err)
					}
				}
				arm("claude-s")
				_, claudeBlock, _ := evaluateStopGoal(ctx, f.root, "claude-s", realCmdRunner{}, nil, os.Stderr)
				claude := codexadapter.DecisionAllow
				if claudeBlock {
					claude = codexadapter.DecisionDeny
				}

				arm("codex-s")
				if tc.receipt != nil {
					key, err := verify.Key(ctx, f.root)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := verify.RecordCheck(f.root, key, verify.CheckEntry{CheckID: "goal", Command: tc.cmd, ExitCode: *tc.receipt, RecordedAt: time.Now()}); err != nil {
						t.Fatal(err)
					}
				}
				c := newCodexStopChain(f.root, stopInput("codex-s", false))
				got := c.goalMember(ctx)
				wantPair(t, claude, got, tc.want, tc.wantClass)
				if tc.name == "receipt absent" && !strings.Contains(got.Reason, codexwiring.GoalReceiptCommand) {
					t.Fatalf("an unmeasured goal must name the receipt command; reason %q", got.Reason)
				}
				if tc.name == "unmet" {
					// The whole Codex chain renders the continuation, not {}.
					c2 := newCodexStopChain(f.root, stopInput("codex-s", false))
					c2.member1 = allowMember1
					c2.budgetFor = wideStopBudget
					out, err := renderCodexStop(c2.run(ctx))
					if err != nil {
						t.Fatal(err)
					}
					if !strings.Contains(string(out), `"decision":"block"`) {
						t.Fatalf("goal-unmet Codex chain output = %s, want a Stop block", out)
					}
				}
			})
		}
	})

	// ── sync gate (member 2) ─────────────────────────────────────────────
	t.Run("sync gate", func(t *testing.T) {
		syncCodex := func(t *testing.T, f *stopFixture, in *hook.HookInput) stopMemberOutcome {
			t.Helper()
			return newCodexStopChain(f.root, in).syncGateMember(ctx)
		}
		produce := func(t *testing.T, f *stopFixture) {
			t.Helper()
			if _, err := produceSyncGateReceipt(ctx, f.root); err != nil {
				t.Fatalf("sync-gate receipt producer: %v", err)
			}
		}

		t.Run("HEAD not a sync-phase commit, receipt absent", func(t *testing.T) {
			f := newStopFixture(t)
			f.write(t, "main.go", "package main\n\nfunc main() { _ = 2 }\n")
			f.commit(t, "feat: ordinary change")
			claude := f.runClaudeSyncGate(t, `{}`)
			got := syncCodex(t, f, stopInput("s", false))
			wantPair(t, claude, got, codexadapter.DecisionAllow, "")
			if got.ReceiptRead {
				t.Fatal("the Codex sync gate read the receipt although HEAD is not a sync-phase commit")
			}
		})
		t.Run("sync-phase commit, failing receipt", func(t *testing.T) {
			f := newStopFixture(t)
			f.setFakeGo(t, 1)
			claude := f.runClaudeSyncGate(t, `{}`)
			produce(t, f)
			got := syncCodex(t, f, stopInput("s", false))
			wantPair(t, claude, got, codexadapter.DecisionDeny, reasonGateFailed)
		})
		t.Run("sync-phase commit, passing receipt", func(t *testing.T) {
			f := newStopFixture(t)
			claude := f.runClaudeSyncGate(t, `{}`)
			produce(t, f)
			got := syncCodex(t, f, stopInput("s", false))
			wantPair(t, claude, got, codexadapter.DecisionAllow, "")
		})
		t.Run("sync-phase commit, receipt absent", func(t *testing.T) {
			f := newStopFixture(t)
			f.setFakeGo(t, 1) // Claude runs the checks in-hook and they fail
			claude := f.runClaudeSyncGate(t, `{}`)
			got := syncCodex(t, f, stopInput("s", false))
			wantPair(t, claude, got, codexadapter.DecisionDeny, reasonUnmeasured)
			if !strings.Contains(got.Reason, codexwiring.SyncGateReceiptCommand) {
				t.Fatalf("an unmeasured sync gate must name its command; reason %q", got.Reason)
			}
		})
		t.Run("sync-phase commit, failing receipt, blocking opt-out", func(t *testing.T) {
			f := newStopFixture(t)
			f.setFakeGo(t, 1)
			t.Setenv("MOAI_SYNC_GATE_BLOCKING", "0")
			claude := f.runClaudeSyncGate(t, `{}`, "MOAI_SYNC_GATE_BLOCKING=0")
			produce(t, f)
			got := syncCodex(t, f, stopInput("s", false))
			wantPair(t, claude, got, codexadapter.DecisionAllow, "")
		})
		t.Run("fresh failing receipt, stop_hook_active true", func(t *testing.T) {
			f := newStopFixture(t)
			f.setFakeGo(t, 1)
			// Claude: the first run stores the fail record and payload; the
			// flagged turn does not re-deliver it, the unflagged one does.
			if d := f.runClaudeSyncGate(t, `{}`); d != codexadapter.DecisionDeny {
				t.Fatalf("premise: the first Claude run must block, got %s", d)
			}
			claude := f.runClaudeSyncGate(t, `{"stop_hook_active": true}`)
			produce(t, f)
			got := syncCodex(t, f, stopInput("s", true))
			wantPair(t, claude, got, codexadapter.DecisionAllow, "")

			claudeAgain := f.runClaudeSyncGate(t, `{"stop_hook_active": false}`)
			gotAgain := syncCodex(t, f, stopInput("s", false))
			wantPair(t, claudeAgain, gotAgain, codexadapter.DecisionDeny, reasonGateFailed)
		})
	})

	// ── codex review gate (member 6) ─────────────────────────────────────
	t.Run("codex review gate", func(t *testing.T) {
		const failReview = "- [P1] found issues\n- [P2] more issues"
		// Post-#1718 parser: prose approval is inconclusive, only a body with
		// the pinned `Verdict: pass` line is a PASS receipt (Opus re-audit F1).
		const passReview = realCleanReview
		setup := func(t *testing.T) *stopFixture {
			t.Helper()
			f := newStopFixture(t)
			f.write(t, "main.go", "package main\n\nfunc main() { _ = 3 }\n")
			f.commit(t, "feat: not a sync commit") // keep member 2 out of the way
			f.enableReviewGates(t, true, false)
			f.dirty(t, "reviewable")
			return f
		}
		claudeReview := func(t *testing.T, f *stopFixture, review string, active bool) codexadapter.Decision {
			t.Helper()
			withCodexSession(t, codexSessionScript(review))
			out, _ := HandleCodexReviewGate(stopInput("s", active), true, f.root)
			return claudeDecision(out)
		}
		produce := func(t *testing.T, f *stopFixture, review string) verify.Receipt {
			t.Helper()
			withCodexSession(t, codexSessionScript(review))
			r, err := produceCodexReviewReceipt(ctx, f.root)
			if err != nil {
				t.Fatalf("codex review receipt producer: %v", err)
			}
			return r
		}
		codex := func(t *testing.T, f *stopFixture, active bool) stopMemberOutcome {
			t.Helper()
			return newCodexStopChain(f.root, stopInput("s", active)).codexReviewMember(ctx)
		}

		t.Run("codex installed, no receipt", func(t *testing.T) {
			f := setup(t)
			claude := claudeReview(t, f, failReview, false)
			got := codex(t, f, false)
			wantPair(t, claude, got, codexadapter.DecisionDeny, reasonUnmeasured)
			if !strings.Contains(got.Reason, codexwiring.CodexReviewReceiptCommand) {
				t.Fatalf("reason must name the review runner; got %q", got.Reason)
			}
		})
		t.Run("codex installed, fresh PASS receipt", func(t *testing.T) {
			f := setup(t)
			claude := claudeReview(t, f, passReview, false)
			r := produce(t, f, passReview)
			// Premise (Opus re-audit F1): this golden's allow leg must be
			// carried by an actual PASS receipt. The post-#1718 parser reads
			// prose approval as inconclusive — which also allows — so without
			// this assertion a fixture regression would silently vacate the
			// "fresh PASS receipt" property while the test stayed green.
			if r.Verdict != codexReviewVerdictPass {
				t.Fatalf("premise: fixture must produce a %q receipt for this leg, got %q", codexReviewVerdictPass, r.Verdict)
			}
			wantPair(t, claude, codex(t, f, false), codexadapter.DecisionAllow, "")
		})
		t.Run("codex installed, fresh INCONCLUSIVE receipt", func(t *testing.T) {
			f := setup(t)
			// Post-#1718 parser: prose approval carries no pinned Verdict
			// line, so the receipt records inconclusive — the review ran but
			// produced no verdict.
			const proseApproval = "Looks good to me — no blocking findings."
			claude := claudeReview(t, f, proseApproval, false)
			if claude != codexadapter.DecisionAllow {
				t.Fatalf("premise: Claude must fail-open an inconclusive review, got %s", claude)
			}
			r := produce(t, f, proseApproval)
			if r.Verdict != codexReviewVerdictInconclusive {
				t.Fatalf("premise: fixture must produce an %q receipt for this leg, got %q", codexReviewVerdictInconclusive, r.Verdict)
			}
			got := codex(t, f, false)
			wantPair(t, claude, got, codexadapter.DecisionAllow, "")
			// t1280 F3: the allow is fail-open, and the record must not read
			// pass for a review that never produced a verdict.
			if got.Status != stopStatusFailOpen {
				t.Fatalf("an inconclusive receipt must be recorded %q, not %q", stopStatusFailOpen, got.Status)
			}
			if len(got.Discards) != 1 || got.Discards[0].Key != "codex-stop-chain/codex-review/inconclusive" {
				t.Fatalf("an inconclusive allow must write exactly one inconclusive discard record, got %+v", got.Discards)
			}
		})
		t.Run("codex installed, FAIL receipt", func(t *testing.T) {
			f := setup(t)
			claude := claudeReview(t, f, failReview, false)
			produce(t, f, failReview)
			wantPair(t, claude, codex(t, f, false), codexadapter.DecisionDeny, reasonGateFailed)
		})
		t.Run("stale receipt, HEAD moved", func(t *testing.T) {
			f := setup(t)
			produce(t, f, passReview)
			f.commit(t, "feat: move HEAD after the review")
			f.dirty(t, "reviewable again")
			claude := claudeReview(t, f, failReview, false)
			got := codex(t, f, false)
			wantPair(t, claude, got, codexadapter.DecisionDeny, reasonUnmeasured)
		})
		t.Run("codex binary missing", func(t *testing.T) {
			f := setup(t)
			withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })
			out, _ := HandleCodexReviewGate(stopInput("s", false), true, f.root)
			got := codex(t, f, false)
			wantPair(t, claudeDecision(out), got, codexadapter.DecisionAllow, "")
			if len(got.Discards) != 1 || !strings.Contains(got.Discards[0].Reason, "codex") {
				t.Fatalf("a missing reviewer must be recorded, got %+v", got.Discards)
			}
		})
		t.Run("no reviewable change", func(t *testing.T) {
			f := setup(t)
			if err := os.Remove(filepath.Join(f.root, "extra.go")); err != nil {
				t.Fatal(err)
			}
			claude := claudeReview(t, f, failReview, false)
			wantPair(t, claude, codex(t, f, false), codexadapter.DecisionAllow, "")
		})
		t.Run("codex installed, no receipt, stop_hook_active true", func(t *testing.T) {
			f := setup(t)
			claude := claudeReview(t, f, failReview, true)
			wantPair(t, claude, codex(t, f, true), codexadapter.DecisionAllow, "")
		})
	})

	// ── multi review gate (member 7) ─────────────────────────────────────
	t.Run("multi review gate", func(t *testing.T) {
		cases := []struct {
			name    string
			verdict string // "" = result missing
			want    codexadapter.Decision
		}{
			{"result blocking", "fail", codexadapter.DecisionDeny},
			{"result passing", "pass", codexadapter.DecisionAllow},
			{"result missing", "", codexadapter.DecisionAllow},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				f := newStopFixture(t)
				f.enableReviewGates(t, false, true)
				f.dirty(t, "reviewable")
				if tc.verdict != "" {
					f.write(t, ".moai/state/audit-multi/s.json", `{"overall_verdict":"`+tc.verdict+`","residual_risk_note":"golden"}`)
				}
				out, _ := HandleMultiReviewGate(stopInput("s", false), true, f.root, "s")
				wantClass := ""
				if tc.want == codexadapter.DecisionDeny {
					wantClass = reasonGateFailed
				}
				got := newCodexStopChain(f.root, stopInput("s", false)).multiReviewMember(ctx)
				wantPair(t, claudeDecision(out), got, tc.want, wantClass)
				if tc.verdict == "" && (len(got.Discards) != 1 || !strings.Contains(got.Discards[0].Reason, "missing")) {
					t.Fatalf("a missing result must be recorded, got %+v", got.Discards)
				}
			})
		}
	})

	// ── Codex-only §D3.8 cap (declared parity deviation; no Claude path) ──
	t.Run("cap", func(t *testing.T) {
		n := codexwiring.StopUnmeasuredCap
		gates := []struct {
			name    string
			member  int
			setup   func(t *testing.T) *stopFixture
			produce func(t *testing.T, f *stopFixture) verify.Receipt
			// invalidate makes the stored receipt stale without moving the key.
			invalidate func(t *testing.T, f *stopFixture)
		}{
			{"sync gate", 2, func(t *testing.T) *stopFixture {
				return newStopFixture(t)
			}, func(t *testing.T, f *stopFixture) verify.Receipt {
				r, err := produceSyncGateReceipt(ctx, f.root)
				if err != nil {
					t.Fatal(err)
				}
				return r
			}, func(t *testing.T, f *stopFixture) {
				// A different `go` on PATH changes the receipt's tool_version,
				// not the tree.
				f.setFakeGo(t, 11)
			}},
			{"codex review gate", 6, func(t *testing.T) *stopFixture {
				f := newStopFixture(t)
				f.write(t, "main.go", "package main\n\nfunc main() { _ = 4 }\n")
				f.commit(t, "feat: not a sync commit")
				f.enableReviewGates(t, true, false)
				f.dirty(t, "reviewable")
				withCodexSession(t, codexSessionScript(realCleanReview))
				return f
			}, func(t *testing.T, f *stopFixture) verify.Receipt {
				withCodexSession(t, codexSessionScript(realCleanReview))
				r, err := produceCodexReviewReceipt(ctx, f.root)
				if err != nil {
					t.Fatal(err)
				}
				return r
			}, func(t *testing.T, _ *stopFixture) {
				// An upgraded reviewer changes the receipt's tool_version, not
				// the tree.
				fakeCodexVersion(t, "codex-cli 0.0.1-golden")
			}},
		}
		for _, g := range gates {
			t.Run(g.name, func(t *testing.T) {
				run := func(t *testing.T, f *stopFixture) stopMemberOutcome {
					t.Helper()
					c := newCodexStopChain(f.root, stopInput("cap-s", false))
					c.member1 = allowMember1
					c.budgetFor = wideStopBudget
					res := c.run(ctx)
					for _, m := range res.Members {
						if m.Number == g.member {
							return m
						}
					}
					t.Fatalf("member %d missing from the chain result", g.member)
					return stopMemberOutcome{}
				}
				t.Run("N-1 consecutive unmeasured Stops continue", func(t *testing.T) {
					f := g.setup(t)
					for i := 1; i < n; i++ {
						if m := run(t, f); m.Decision != codexadapter.DecisionDeny || m.Class != reasonUnmeasured {
							t.Fatalf("Stop %d of %d: got %s/%q, want a continuation (unmeasured)", i, n-1, m.Decision, m.Class)
						}
					}
				})
				t.Run("the Nth Stop allows and records unverified", func(t *testing.T) {
					f := g.setup(t)
					for i := 1; i < n; i++ {
						run(t, f)
					}
					before := countDiscards(sinkRecords(t, f.root), "unverified")
					m := run(t, f)
					if m.Decision != codexadapter.DecisionAllow || m.Class != reasonUnverified {
						t.Fatalf("Stop %d: got %s/%q, want allow/unverified", n, m.Decision, m.Class)
					}
					if after := countDiscards(sinkRecords(t, f.root), "unverified"); after-before != 1 {
						t.Fatalf("the cap must write exactly one unverified record, wrote %d", after-before)
					}
					if !strings.Contains(m.Reason, "unverified") {
						t.Fatalf("the cap reason must name the gate unverified: %q", m.Reason)
					}
					rec, err := readCodexStopChainRecord(f.root, "cap-s")
					if err != nil {
						t.Fatal(err)
					}
					if st := rec.status(g.member); st != reasonUnverified {
						t.Fatalf("verdict record reads %q for member %d, want %q (never a pass)", st, g.member, reasonUnverified)
					}
				})
				t.Run("a fresh receipt resets the count", func(t *testing.T) {
					f := g.setup(t)
					for i := 1; i < n; i++ {
						run(t, f)
					}
					// Premise (delta re-audit D2): the count reset must be
					// carried by an actual PASS receipt — an inconclusive
					// one also allows, and would leave this subtest green
					// without the pass property (same shape as F1).
					if r := g.produce(t, f); r.Verdict != "pass" {
						t.Fatalf("premise: the fresh receipt must be a pass receipt, got %q", r.Verdict)
					}
					if m := run(t, f); m.Decision != codexadapter.DecisionAllow || m.Status != stopStatusPass {
						t.Fatalf("with a fresh receipt: got %s/%q, want an evaluated pass", m.Decision, m.Status)
					}
					// Invalidate the receipt WITHOUT moving HEAD or the working-tree
					// digest, so only the receipt read — not a key change — can have
					// reset the count.
					keyBefore, err := verify.Key(ctx, f.root)
					if err != nil {
						t.Fatal(err)
					}
					g.invalidate(t, f)
					if keyAfter, _ := verify.Key(ctx, f.root); keyAfter != keyBefore {
						t.Fatalf("premise: the invalidation moved the tree key (%s → %s)", keyBefore, keyAfter)
					}
					if m := run(t, f); m.Decision != codexadapter.DecisionDeny || m.Class != reasonUnmeasured {
						t.Fatalf("after a fresh receipt: got %s/%q, want a continuation (count reset to 1)", m.Decision, m.Class)
					}
				})
				t.Run("a HEAD change resets the count", func(t *testing.T) {
					f := g.setup(t)
					for i := 1; i < n; i++ {
						run(t, f)
					}
					f.write(t, "main.go", "package main\n\nfunc main() { _ = 7 }\n")
					if g.member == 2 {
						f.commit(t, "docs(SPEC-FX-001): sync-phase close again")
					} else {
						f.commit(t, "feat: HEAD moved")
						f.dirty(t, "reviewable after the move")
					}
					if m := run(t, f); m.Decision != codexadapter.DecisionDeny || m.Class != reasonUnmeasured {
						t.Fatalf("after a HEAD change: got %s/%q, want a continuation (count reset to 1)", m.Decision, m.Class)
					}
				})
			})
		}
	})
}

func intp(n int) *int { return &n }

// allowMember1 stands in for `moai hook stop` where a golden isolates another
// member.
func allowMember1(context.Context, *hook.HookInput) (*hook.HookOutput, error) {
	return &hook.HookOutput{}, nil
}
