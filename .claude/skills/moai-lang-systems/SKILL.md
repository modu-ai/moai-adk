---
name: moai-lang-systems
description: "Systems languages consolidated: C, C++, Shell, SQL, R, Ruby, PHP for scripting and systems"
version: 1.0.0
modularized: false
last_updated: 2025-11-24
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - Read
  - Write
  - Edit
compliance_score: 80
modules: []
dependencies:
  - moai-foundation-trust
deprecated: false
successor: null
category_tier: 4
auto_trigger_keywords:
  - c
  - cpp
  - shell
  - sql
  - ruby
  - php
  - script
  - bash
  - query
  - database
  - systems
agent_coverage:
  - backend-expert
  - database-expert
  - tdd-implementer
  - quality-gate
context7_references:
  - c
  - cpp
  - sqlite
invocation_api_version: "1.0"
---

## Quick Reference (30 seconds)

**Systems Languages Consolidated**

Unified collection of systems programming, scripting, and database query languages: C, C++, Shell, SQL, R, Ruby, PHP.

**When to Use**:
- Systems programming with C/C++
- DevOps automation with Shell/Bash
- Database operations with SQL
- Data analysis with R
- Web development with Ruby/PHP
- Legacy system maintenance

---

## Core Languages

### C/C++

Low-level systems programming with maximum performance.

```c
// C: Simple, fast
#include <stdio.h>

int main() {
    printf("Hello, World!\n");
    return 0;
}
```

### Shell/Bash

DevOps and system automation.

```bash
#!/bin/bash
# Deploy application
docker build -t myapp . && docker run myapp
```

### SQL

Database queries and data manipulation.

```sql
-- Query users with posts
SELECT u.*, COUNT(p.id) as post_count
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
GROUP BY u.id
HAVING COUNT(p.id) > 0;
```

### Ruby/PHP

Web development and scripting.

```ruby
# Ruby web framework example
class User
  attr_accessor :name, :email

  def initialize(name, email)
    @name = name
    @email = email
  end
end
```

---

**Status**: Production Ready
**Generated with**: MoAI-ADK Skill Factory
