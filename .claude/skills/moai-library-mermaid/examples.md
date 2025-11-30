# Example Diagrams - All 21 Types

Quick reference examples for each Mermaid diagram type.

## 1. Flowchart Example

```mermaid
flowchart TD
 Start([Start]) --> Input[Input Data]
 Input --> Process{Validation}
 Process -->|Valid| Calculate[Calculate Result]
 Process -->|Invalid| Error[Error Handler]
 Calculate --> Output[Output]
 Error --> Output
 Output --> End([End])
 
 style Start fill:#90EE90
 style End fill:#FFB6C6
 style Error fill:#FFB6C6
```

## 2. Sequence Diagram Example

```mermaid
sequenceDiagram
 participant User
 participant API
 participant Database
 
 User->>API: GET /users
 API->>Database: SELECT * FROM users
 Database-->>API: User records
 API-->>User: JSON response
```

## 3. Class Diagram Example

```mermaid
classDiagram
 class Animal {
 string name
 eat()
 }
 class Dog {
 bark()
 }
 Animal <|-- Dog
```

## 4. State Diagram Example

```mermaid
stateDiagram-v2
 [*] --> Idle
 Idle --> Working: start
 Working --> Idle: complete
 Working --> Error: fail
 Error --> Idle: reset
 Idle --> [*]
```

## 5. ER Diagram Example

```mermaid
erDiagram
 CUSTOMER ||--o{ ORDER : places
 ORDER ||--|{ ORDER-ITEM : contains
 PRODUCT ||--o{ ORDER-ITEM : "included in"
 
 CUSTOMER {
 int id PK
 string email
 }
```

## 6. Gantt Chart Example

```mermaid
gantt
 title Project Schedule
 dateFormat YYYY-MM-DD
 
 section Phase 1
 Design :des, 2025-01-01, 30d
 Implementation :impl, after:des, 45d
 Testing :test, after:impl, 20d
```

## 7. Mindmap Example

```mermaid
mindmap
 root((Project))
 Planning
 Timeline
 Resources
 Execution
 Development
 Testing
 Delivery
 Deployment
 Documentation
```

## 8. Timeline Example

```mermaid
timeline
 title Development Timeline
 2025-01 : Project Start : Team Assembled
 2025-02 : Development
 2025-03 : Testing
 2025-04 : Launch
```

## 9. Git Graph Example

```mermaid
gitGraph
 commit id: "Initial"
 branch develop
 commit id: "Feature 1"
 checkout main
 merge develop
 commit id: "v1.0"
```

## 10. C4 Diagram Example

```mermaid
C4Context
 title System Context
 Person(user, "User", "Needs service")
 System(system, "System", "Provides service")
 Rel(user, system, "Uses")
```

## 11. User Journey Example

```mermaid
journey
 title Onboarding
 section Sign Up
 Form : 4 : User
 Verify : 3 : User
 section Setup
 Profile : 5 : User
```

## 12. Requirement Diagram Example

```mermaid
requirementDiagram
 requirement REQ1 {
 id: 1
 text: Authentication
 risk: High
 }
 element webapp {
 type: Software
 }
 REQ1 - satisfies - webapp
```

## 13. Pie Chart Example

```mermaid
pie title Stack Distribution
 "Frontend" : 40
 "Backend" : 35
 "DevOps" : 25
```

## 14. Quadrant Chart Example

```mermaid
quadrantChart
 title Priority Matrix
 x-axis Low --> High
 y-axis Low --> High
 Task A: [0.2, 0.8]
 Task B: [0.8, 0.2]
```

## 15. XY Chart Example

```mermaid
xychart-beta
 title Performance
 x-axis [Jan, Feb, Mar]
 y-axis "Response Time" 0 --> 300
 line [150, 160, 145]
```

## 16. Block Diagram Example

```mermaid
block-beta
 columns 3
 A["Input"]
 B["Process"]
 C["Output"]
 A --> B --> C
```

## 17. Kanban Example

```mermaid
kanban
 section Todo
 Task 1
 Task 2
 
 section Doing
 Task 3
 
 section Done
 Task 4
```

## 18. Sankey Diagram Example

```mermaid
sankey-beta
 Users, 100
 Users, Development, 60
 Users, QA, 40
 Development, Production, 60
 QA, Production, 40
```

## 19. Packet Diagram Example

```mermaid
packet-beta
 0-7: Source Port
 8-15: Dest Port
 16-31: Sequence
 32-47: ACK
```

## 20. Radar Chart Example

```mermaid
radar
 title Skills
 Frontend: 90
 Backend: 85
 DevOps: 75
 Design: 70
 max: 100
```

## 21. Architecture Diagram Example

```mermaid
graph TB
 Web["Web App"]
 API["API Server"]
 DB["Database"]
 Cache["Redis"]
 
 Web --> API
 API --> DB
 API --> Cache
```

---

For more examples, visit [mermaid.live](https://mermaid.live)
