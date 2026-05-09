#!/usr/bin/env node
import { runAdapter } from "./moai-teams-hook-adapter.mjs";

try {
  process.exitCode = await runAdapter("task_completed");
} catch (err) {
  console.error(err instanceof Error ? err.message : String(err));
  process.exitCode = 1;
}
