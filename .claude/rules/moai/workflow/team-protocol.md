# Team Protocol

Shared protocol for all MoAI Agent Teams teammates. This supplements role-specific instructions in each agent definition and spawn prompt.

## Team Discovery

- Read `~/.claude/teams/{team-name}/config.json` to discover teammates
- Always refer to teammates by their `name` field when using SendMessage

## Communication

- Use direct messages (type: "message") by default
- NEVER broadcast unless a critical blocking issue affects ALL teammates
- Send findings and results to the team lead via SendMessage when complete
- Report blockers to the team lead immediately
- Update task status via TaskUpdate

## Task Management

After completing each task:
- Mark task as completed via TaskUpdate (MANDATORY - prevents infinite waiting)
- Check TaskList for available unblocked tasks
- Claim the next available unblocked task (prefer lowest ID first) or wait for team lead instructions

## Error Recovery

- If you encounter an error, do NOT stop working. Try an alternative approach first
- If the error persists after 3 attempts, report it to the team lead via SendMessage with the error details, file path, and what you tried
- Continue with remaining tasks even if one task fails
- If blocked by another teammate's work, report the blocker and move to the next unblocked task

## Shutdown Handling

When you receive a shutdown_request JSON message:
- If all work is complete: SendMessage(type: "shutdown_response", request_id: "<from message>", approve: true)
- If work is in progress: SendMessage(type: "shutdown_response", request_id: "<from message>", approve: false, content: "Still working on [task]")

## Idle States

- Going idle is NORMAL - it means you are waiting for input from the team lead
- After completing work, you will go idle while waiting for the next assignment
- The team lead will either send new work or a shutdown request
- NEVER assume work is done until you receive shutdown_request from the lead

## Context Isolation

- You do NOT have access to the team lead's conversation history
- All necessary context must come from your spawn prompt or teammate messages
- If context is insufficient, ask the team lead for clarification via SendMessage
