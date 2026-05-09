import type { ClaudeTeamApi } from "./team-schema.ts";

export interface NormalizedTeamCall {
  api: ClaudeTeamApi;
  params: Record<string, unknown>;
}

export interface TmustierTeamsToolCall {
  tool: "teams";
  params: Record<string, unknown>;
}

export function toTmustierTeamsToolCall(call: NormalizedTeamCall): TmustierTeamsToolCall {
  const p = call.params;
  switch (call.api) {
    case "TeamCreate":
      return {
        tool: "teams",
        params: {
          action: "delegate",
          teammates: p.teammates,
          tasks: p.tasks,
          contextMode: p.contextMode ?? "branch",
          workspaceMode: p.workspaceMode ?? "worktree",
          model: p.model,
          thinking: p.thinking,
          planRequired: p.planRequired,
        },
      };
    case "SendMessage":
      return {
        tool: "teams",
        params: p.name
          ? { action: p.urgent ? "message_steer" : "message_dm", name: p.name, message: p.message, urgent: p.urgent }
          : { action: "message_broadcast", message: p.message, urgent: p.urgent },
      };
    case "TaskCreate":
      return { tool: "teams", params: { action: "delegate", tasks: p.tasks ?? [{ text: p.text, assignee: p.assignee }] } };
    case "TaskUpdate":
      if (p.status) return { tool: "teams", params: { action: "task_set_status", taskId: p.taskId, status: p.status } };
      if (p.assignee) return { tool: "teams", params: { action: "task_assign", taskId: p.taskId, assignee: p.assignee } };
      return { tool: "teams", params: { action: "task_unassign", taskId: p.taskId } };
    case "TaskList":
      return { tool: "teams", params: { action: "member_status" } };
    case "TaskGet":
      return { tool: "teams", params: { action: "task_dep_ls", taskId: p.taskId } };
    case "TeamDelete":
      return { tool: "teams", params: { action: "team_done", all: p.force ?? true } };
  }
}

export function teamRuntimeStatus(): string {
  return "ok: @tmustier/pi-agent-teams normalized Team API mapping implemented for teams tool; live teammate invocation pending";
}
