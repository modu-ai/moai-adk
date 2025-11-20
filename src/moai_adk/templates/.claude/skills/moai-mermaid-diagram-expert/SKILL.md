---
name: moai-mermaid-diagram-expert
version: 5.0.0
status: stable
description: Enterprise Mermaid diagramming solution with 21 diagram types
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# Mermaid Diagram Expert

**Comprehensive Mermaid.js solution for visualization and documentation**

> **Version**: 5.0.0  
> **Focus**: Flowcharts, Sequence, Class, ER, Gantt, C4, Architecture  
> **Tools**: Mermaid CLI, Python Converter

---

## Quick Start

### Installation

```bash
# Install Mermaid CLI
npm install -g @mermaid-js/mermaid-cli

# Convert diagram
mmdc -i diagram.mmd -o output.png
```

### Python Converter (Custom)

```python
# scripts/mermaid-to-svg-png.py
import sys
import subprocess

def convert(input_file, output_file, format='png'):
    cmd = ['mmdc', '-i', input_file, '-o', output_file]
    if format == 'png':
        cmd.extend(['-b', 'transparent'])
    subprocess.run(cmd, check=True)

if __name__ == '__main__':
    convert(sys.argv[1], sys.argv[2])
```

---

## Core Diagram Types

### 1. Flowchart

Process flows and decision trees.

```mermaid
flowchart TD
    Start([Start]) --> Input[Input Data]
    Input --> Process{Valid?}
    Process -->|Yes| Calculate[Calculate]
    Process -->|No| Error[Handle Error]
    Calculate --> Output[Output Result]
    Error --> Output
    Output --> End([End])

    style Start fill:#90EE90
    style End fill:#FFB6C6
```

### 2. Sequence Diagram

Actor interactions and API flows.

```mermaid
sequenceDiagram
    participant User
    participant API
    participant DB

    User->>API: POST /login
    API->>DB: Validate User
    DB-->>API: User Record
    API-->>User: Token

    alt Success
        User->>API: GET /data
        API-->>User: Data
    else Error
        User->>API: GET /data
        API-->>User: 401 Unauthorized
    end
```

### 3. Class Diagram

OOP architecture and data modeling.

```mermaid
classDiagram
    class User {
        +String name
        +String email
        +login()
        +logout()
    }

    class Order {
        +int id
        +Date date
        +process()
    }

    User "1" --> "*" Order : places
```

### 4. Entity Relationship (ER)

Database schema and relationships.

```mermaid
erDiagram
    USER ||--o{ ORDER : places
    ORDER ||--|{ ORDER_ITEM : contains
    PRODUCT ||--o{ ORDER_ITEM : "included in"

    USER {
        int id PK
        string email
    }

    ORDER {
        int id PK
        int user_id FK
    }
```

### 5. Gantt Chart

Project timelines and schedules.

```mermaid
gantt
    title Project Timeline
    dateFormat YYYY-MM-DD

    section Phase 1
    Design :a1, 2024-01-01, 30d
    Dev    :after a1, 60d

    section Phase 2
    Review :2024-04-01, 15d
    Deploy :milestone, 2024-04-15, 0d
```

### 6. State Diagram

State machines and lifecycle.

```mermaid
stateDiagram-v2
    [*] --> Idle
    Idle --> Processing : Start
    Processing --> Success : Complete
    Processing --> Error : Fail
    Error --> Idle : Retry
    Success --> [*]
```

### 7. C4 Architecture

System context and container diagrams.

```mermaid
C4Context
    title System Context
    Person(user, "User", "Uses system")
    System(system, "System", "Core functionality")
    System_Ext(mail, "Email System", "Sends notifications")

    Rel(user, system, "Uses")
    Rel(system, mail, "Sends emails")
```

### 8. Git Graph

Version control history.

```mermaid
gitGraph
    commit
    branch develop
    checkout develop
    commit
    commit
    checkout main
    merge develop
    commit tag: "v1.0"
```

### 9. Mindmap

Hierarchical knowledge structure.

```mermaid
mindmap
  root((Project))
    Planning
      Requirements
      Timeline
    Development
      Frontend
      Backend
    Deployment
      CI/CD
      Monitoring
```

### 10. User Journey

User experience flows.

```mermaid
journey
    title User Onboarding
    section Sign Up
      Register: 5: User
      Verify Email: 3: User
    section Setup
      Profile: 4: User
      First Project: 5: User
```

---

## Advanced Features

### Styling & Theming

```mermaid
%%{init: {'theme': 'base', 'themeVariables': { 'primaryColor': '#ff0000'}}}%%
graph TD
    A[Custom Color]
```

### Subgraphs

```mermaid
graph TB
    subgraph Backend
        API --> DB
    end
    subgraph Frontend
        UI --> API
    end
```

---

## Best Practices

1.  **Keep it Simple**: Limit nodes to <20 for readability.
2.  **Direction**: Use `TD` (Top-Down) for hierarchies, `LR` (Left-Right) for flows.
3.  **Styling**: Use classes for consistent styling across nodes.
4.  **Version Control**: Commit `.mmd` files, generate images in CI.

---

## Validation Checklist

- [ ] Syntax valid (renders in editor)
- [ ] Direction appropriate for content
- [ ] Labels clear and concise
- [ ] Colors used for semantic meaning
- [ ] Complex logic broken into subgraphs

---

## Related Skills

- `moai-project-documentation`: Documentation standards
- `moai-domain-frontend`: UI visualization
- `moai-domain-backend`: API & DB modeling

---

**Last Updated**: 2025-11-20
