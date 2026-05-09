import { TEAM_BACKEND_PRIORITY } from "./constants.ts";

export type ClaudeTeamApi = "TeamCreate" | "SendMessage" | "TaskCreate" | "TaskUpdate" | "TaskList" | "TaskGet" | "TeamDelete";

export interface TeamApiDescriptor {
  api: ClaudeTeamApi;
  primary: string;
  primaryAction: string;
  fallback: string[];
  adapterStatus: "schema-represented" | "runtime-pending";
  normalizedPurpose: string;
}

const PURPOSES: Record<ClaudeTeamApi, string> = {
  TeamCreate: "Create a named team with role/profile/worktree policy and launch teammates.",
  SendMessage: "Send direct or broadcast message to teammate(s).",
  TaskCreate: "Create shared team task with dependencies/owner/status.",
  TaskUpdate: "Update task status, owner, progress, or completion metadata.",
  TaskList: "List tasks by team/status/dependency readiness.",
  TaskGet: "Inspect one task in detail.",
  TeamDelete: "Gracefully finish and cleanup team resources.",
};

const TMUSTIER_ACTIONS: Record<ClaudeTeamApi, string> = {
  TeamCreate: "delegate or member_spawn",
  SendMessage: "message_dm, message_broadcast, or message_steer",
  TaskCreate: "delegate with tasks or package task add command",
  TaskUpdate: "task_assign/task_unassign/task_set_status/task_dep_add/task_dep_rm",
  TaskList: "member_status plus task board query command",
  TaskGet: "task_dep_ls or task board detail command",
  TeamDelete: "team_done, member_shutdown, member_kill, member_prune",
};

export function getTeamSchemaDescriptors(): TeamApiDescriptor[] {
  const [primary, ...fallback] = TEAM_BACKEND_PRIORITY;
  return (Object.keys(PURPOSES) as ClaudeTeamApi[]).map((api) => ({
    api,
    primary,
    primaryAction: TMUSTIER_ACTIONS[api],
    fallback,
    adapterStatus: "runtime-pending",
    normalizedPurpose: PURPOSES[api],
  }));
}

export function formatTeamSchemaReport(): string[] {
  return [
    "Agent Teams schema spike: represented from @tmustier/pi-agent-teams README; runtime adapter pending",
    ...getTeamSchemaDescriptors().map(
      (d) => `pending: ${d.api} -> ${d.primary} action '${d.primaryAction}'; fallback ${d.fallback.join(" > ")}`,
    ),
  ];
}
