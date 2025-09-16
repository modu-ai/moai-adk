# Interactive Mode

## Keyboard Shortcuts

### General Controls

| Shortcut       | Description                        | Context                    |
| -------------- | ---------------------------------- | -------------------------- |
| `Ctrl+C`       | Cancel current input or generation | Standard interrupt         |
| `Ctrl+D`       | Exit Claude Code session           | EOF signal                 |
| `Ctrl+L`       | Clear terminal screen              | Keeps conversation history |
| Up/Down arrows | Navigate command history           | Recall previous inputs     |
| `Esc` + `Esc`  | Edit previous message              | Double-escape to modify    |

### Multiline Input

| Method           | Shortcut       | Context                           |
| ---------------- | -------------- | --------------------------------- |
| Quick escape     | `\` + `Enter`  | Works in all terminals            |
| macOS default    | `Option+Enter` | Default on macOS                  |
| Terminal setup   | `Shift+Enter`  | After `/terminal-setup`           |
| Control sequence | `Ctrl+J`       | Line feed character for multiline |
| Paste mode       | Paste directly | For code blocks, logs             |

### Quick Commands

| Shortcut     | Description                        | Notes                      |
| ------------ | ---------------------------------- | -------------------------- |
| `#` at start | Memory shortcut - add to CLAUDE.md | Prompts for file selection |
| `/` at start | Slash command                      | See slash commands         |

## Vim Mode

Enable vim-style editing with `/vim` command or configure permanently via `/config`.

### Mode Switching

| Command | Action                      | From Mode |
| ------- | --------------------------- | --------- |
| `Esc`   | Enter NORMAL mode           | INSERT    |
| `i`     | Insert before cursor        | NORMAL    |
| `I`     | Insert at beginning of line | NORMAL    |
| `a`     | Insert after cursor         | NORMAL    |
| `A`     | Insert at end of line       | NORMAL    |
| `o`     | Open line below             | NORMAL    |
| `O`     | Open line above             | NORMAL    |

### Navigation (NORMAL mode)

| Command         | Action                  |
| --------------- | ----------------------- |
| `h`/`j`/`k`/`l` | Move left/down/up/right |
| `w`             | Next word               |
| `b`             | Previous word           |
| `0`             | Beginning of line       |
| `$`             | End of line             |
| `gg`            | Beginning of input      |
| `G`             | End of input            |

### Text Operations (NORMAL mode)

| Command  | Action              |
| -------- | ------------------- |
| `x`      | Delete character    |
| `dd`     | Delete line         |
| `yy`     | Yank (copy) line    |
| `p`      | Paste after cursor  |
| `P`      | Paste before cursor |
| `u`      | Undo                |
| `Ctrl+R` | Redo                |

## Conversation Management

### History Navigation

- Use up/down arrows to navigate through previous commands
- Search history with `Ctrl+R` in some terminals

### Context Control

- `/clear` - Clear conversation history
- `/compact` - Compress conversation to save context
- Resume sessions with `claude --resume`

## Tips for Efficient Use

1. **Quick Edits**: Double-escape to edit previous message
2. **Batch Commands**: Send multiple instructions in one message
3. **Memory Shortcuts**: Use `#` prefix to quickly add to CLAUDE.md
4. **Vim Power**: Enable vim mode for advanced text editing
5. **Multiline Code**: Use `\` + Enter for easy multiline input
