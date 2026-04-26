# Mermaid Diagrams Guide

All documentation diagrams use Mermaid and support light/dark mode automatically.

## Supported Diagram Types

### 1. Flowchart (Used Most Often)

Shows process flows and decision trees:

```mermaid
graph TD
    A["📥 Input"] --> B["🔧 Process"]
    B --> C{Decision}
    C -->|Yes| D["✅ Success"]
    C -->|No| E["❌ Failed"]
    
    style A fill:#4F46E5,stroke:#312E81,color:#fff
    style D fill:#10B981,stroke:#065F46,color:#fff
    style E fill:#EF4444,stroke:#7F1D1D,color:#fff
```

**Best for**:
- Optimization pipeline flows
- Request processing paths
- Decision trees
- User workflows

### 2. Sequence Diagram

Shows interaction between components:

```mermaid
sequenceDiagram
    participant User
    participant Gateway
    participant Cache
    participant Claude
    
    User ->> Gateway: Send request
    Gateway ->> Cache: Check cache
    Cache -->> Gateway: Cache hit!
    Gateway -->> User: Return cached response
```

**Best for**:
- Request/response flows
- Component interactions
- System communication

### 3. Graph/Node Diagram

Shows relationships and hierarchies:

```mermaid
graph LR
    A["authenticate()"] --> B["login()"]
    A --> C["verify_user()"]
    B --> D["User"]
    C --> D
    
    style A fill:#3B82F6,stroke:#1E40AF,color:#fff
    style D fill:#10B981,stroke:#065F46,color:#fff
```

**Best for**:
- Knowledge graph relationships
- Class hierarchies
- Dependency trees
- Code relationships

## Color Scheme

All diagrams use CSS variables for automatic light/dark compatibility:

| Use Case | Color | Hex | When |
|----------|-------|-----|------|
| Input/Start | Primary Blue | #4F46E5 | Initial state, inputs |
| Success/Output | Green | #10B981 | Positive results, cache hits |
| Error/Blocked | Red | #EF4444 | Failures, security blocks |
| Processing/Steps | Amber | #F59E0B | Intermediate steps, layers |
| Info/Details | Cyan | #3B82F6 | Additional information |
| Secondary/Alternative | Purple | #8B5CF6 | Alternative paths |

## Examples by Document

### README.md

**Optimization Pipeline** (7-layer flow):
```mermaid
graph TD
    A["Request"] --> B["Cache Bypass Check"]
    B -->|--no-cache| C["Fresh Response"]
    B -->|Normal| D["Exact Dedup"]
    D -->|Hit| E["Return Cached"]
    D -->|Miss| F["Graph Lookup"]
    F -->|Hit| G["Return Graph Result"]
    F -->|Miss| H["Input Optimization"]
    H --> I["Claude API"]
    
    style A fill:#4F46E5,stroke:#312E81,color:#fff
    style E fill:#10B981,stroke:#065F46,color:#fff
    style G fill:#10B981,stroke:#065F46,color:#fff
    style C fill:#10B981,stroke:#065F46,color:#fff
```

### KNOWLEDGE_GRAPH.md

**Indexing Pipeline** (file → graph):
```mermaid
graph LR
    A["Source Files"] --> B["File Watcher"]
    B --> C["AST Parser"]
    C --> D["Relationship Extract"]
    D --> E["SQLite Storage"]
    E --> F["Query Ready"]
```

**Query Flow** (question → answer):
```mermaid
graph TD
    A["User Query"] --> B["Graph Answerable?"]
    B -->|Yes| C["Graph Lookup"]
    B -->|No| D["Claude API"]
    C --> E["Fast Result"]
    D --> E
```

### INPUT_OPTIMIZATION.md

**Optimization Layers** (input → output):
```mermaid
graph TD
    A["Input 1330t"] --> B["Tool Strip<br/>-35%"]
    B --> C["Param Compress<br/>-60%"]
    C --> D["Format<br/>-60%"]
    D --> E["Whitespace<br/>-37%"]
    E --> F["Output 800t<br/>40% saved"]
```

## Light/Dark Mode Support

All diagrams automatically adapt to system theme:

```html
<!-- In documentation HTML head -->
<style>
  @media (prefers-color-scheme: dark) {
    :root {
      --mermaid-primary: #6366F1;      /* Lighter blue for dark mode */
      --mermaid-success: #34D399;      /* Lighter green for dark mode */
      /* etc */
    }
  }
</style>

<!-- Mermaid will use these CSS variables -->
<div class="mermaid">
  graph TD
    A["Input"] --> B["Process"]
    style A fill:var(--mermaid-primary),stroke:var(--mermaid-primary-dark),color:#fff
</div>
```

## Rendering

### GitHub

Mermaid diagrams render automatically in GitHub markdown:

```markdown
# My Diagram

```mermaid
graph TD
  A --> B
```

No special setup needed.
```

### Local/HTML

Mermaid requires a script tag:

```html
<script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
<script>
  mermaid.initialize({ startOnLoad: true, theme: 'default' });
</script>

<div class="mermaid">
  graph TD
    A --> B
</div>
```

### VS Code

Install "Markdown Preview Mermaid Support" extension:

```bash
code --install-extension bierner.markdown-mermaid
```

Then view diagrams in preview pane.

## Best Practices

### 1. Use Descriptive Labels

```mermaid
graph TD
    A["❌ Bad: Generic label"] --> B
    C["✅ Good: Specific action<br/>with icon"] --> D
    
    style A fill:#EF4444,stroke:#7F1D1D,color:#fff
    style C fill:#10B981,stroke:#065F46,color:#fff
```

### 2. Add Emojis for Quick Recognition

- 📥 Inputs/Data
- 🔧 Processing/Tools
- 💾 Storage/Cache
- 🤖 AI/Claude
- ✅ Success
- ❌ Error
- ⚡ Fast/Efficient
- 🔒 Security

### 3. Use Consistent Styling

Always use CSS variables for colors, not hardcoded hex:

```mermaid
graph TD
    A["Use Variables"] 
    style A fill:var(--mermaid-primary),stroke:var(--mermaid-primary-dark),color:#fff
    
    B["NOT: Hardcoded"]
    style B fill:#4F46E5,stroke:#312E81,color:#fff
```

### 4. Keep Diagrams Simple

If a diagram needs many nodes (>15), break it into multiple diagrams:

```mermaid
graph TD
    A["Overview"] --> B["Detail 1"]
    A --> C["Detail 2"]
    
    subgraph detail1["Detail 1 Flow"]
      D["Step 1"] --> E["Step 2"]
    end
    subgraph detail2["Detail 2 Flow"]
      F["Step 1"] --> G["Step 2"]
    end
    
    B --> detail1
    C --> detail2
```

### 5. Show Data Flow

Include token counts, timing, or resource usage:

```mermaid
graph LR
    A["Request<br/>1330 tokens"] --> B["Processing<br/>+50ms"]
    B --> C["Response<br/>800 tokens<br/>40% saved"]
    
    style A fill:#EF4444,stroke:#7F1D1D,color:#fff
    style C fill:#10B981,stroke:#065F46,color:#fff
```

## Common Patterns

### Decision Point

```mermaid
graph TD
    A["Check Condition"] --> B{Is True?}
    B -->|Yes| C["Path 1"]
    B -->|No| D["Path 2"]
    
    style A fill:#4F46E5,stroke:#312E81,color:#fff
    style C fill:#10B981,stroke:#065F46,color:#fff
    style D fill:#F59E0B,stroke:#92400E,color:#fff
```

### Parallel Paths

```mermaid
graph TD
    A["Start"] --> B["Path 1"]
    A --> C["Path 2"]
    A --> D["Path 3"]
    B --> E["Merge"]
    C --> E
    D --> E
    
    style A fill:#4F46E5,stroke:#312E81,color:#fff
    style E fill:#10B981,stroke:#065F46,color:#fff
```

### Feedback Loop

```mermaid
graph TD
    A["Input"] --> B["Process"]
    B --> C["Output"]
    C --> D{Good?}
    D -->|No| B
    D -->|Yes| E["Done"]
    
    style A fill:#4F46E5,stroke:#312E81,color:#fff
    style E fill:#10B981,stroke:#065F46,color:#fff
```

## Testing Diagrams Locally

View mermaid rendering in VS Code:

1. Open markdown file
2. Press `Ctrl+K V` (or `Cmd+K V` on Mac)
3. Mermaid diagrams render in preview pane
4. System dark/light mode preference applies automatically

## Updating Existing Diagrams

When updating ASCII diagrams to Mermaid:

1. **Keep the information** — Mermaid should show the same flow
2. **Add visual hierarchy** — Color code by function (input, process, output)
3. **Use emojis** — Quick visual recognition
4. **Test in dark mode** — Ensure contrast is sufficient
5. **Simplify if needed** — Break large diagrams into smaller ones

## References

- [Mermaid Documentation](https://mermaid.js.org/)
- [Mermaid Live Editor](https://mermaid.live/) — Test diagrams online
- [CSS Variables Reference](https://developer.mozilla.org/en-US/docs/Web/CSS/--*) — For theming
