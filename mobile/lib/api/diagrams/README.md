# PlantUML Diagrams

This directory contains PlantUML diagrams for the API client architecture and flows.

## Diagrams

### 1. Architecture Overview
**File:** `architecture.puml`
**Type:** Component diagram
**Description:** Shows the overall architecture of the API layer, including the relationship between `client.ts`, `auth.ts`, `api.ts`, and components.

### 2. Successful Request Flow
**File:** `successful-request-flow.puml`
**Type:** Sequence diagram
**Description:** Shows the normal flow of a successful API request with valid token, demonstrating how `authMiddleware` adds the Authorization header.

### 3. Token Refresh Flow
**File:** `token-refresh-flow.puml`
**Type:** Sequence diagram
**Description:** Shows the automatic token refresh flow when a 401 error is detected, including:
- How `refreshMiddleware` detects 401
- Singleton pattern to prevent concurrent refreshes
- Token refresh using `baseClient` without middleware
- Retry with native `fetch()` using the new token

### 4. Failed Refresh Flow
**File:** `failed-refresh-flow.puml`
**Type:** Sequence diagram
**Description:** Shows what happens when token refresh fails (both access and refresh tokens expired), including token cleanup and redirect to login.

### 5. Infinite Recursion Problem
**File:** `infinite-recursion-problem.puml`
**Type:** Sequence diagram
**Description:** Demonstrates the problem that occurs when using a single client with middleware for both auth operations and protected endpoints (infinite recursion).

### 6. Two Clients Solution
**File:** `two-clients-solution.puml`
**Type:** Component diagram
**Description:** Compares the wrong approach (one client with middleware) versus the correct approach (two separate clients).

### 7. Middleware Execution
**File:** `middleware-execution.puml`
**Type:** Activity diagram
**Description:** Shows the complete middleware execution flow, including decision points for unprotected routes, token validation, and refresh logic.

## Rendering PlantUML Diagrams

**Note:** All diagrams use PlantUML's built-in renderer and **do not require Graphviz installation**.

### VS Code
Install the [PlantUML extension](https://marketplace.visualstudio.com/items?itemName=jebbs.plantuml) and press `Alt+D` to preview diagrams.

No additional setup needed - all diagrams work with the built-in PlantUML renderer.

### Online
Use [PlantUML Online Editor](https://www.plantuml.com/plantuml/uml/) to render diagrams.

### Command Line
```bash
# Install PlantUML (Java required)
brew install plantuml  # macOS
apt-get install plantuml  # Ubuntu/Debian

# Or download JAR directly
wget https://github.com/plantuml/plantuml/releases/download/v1.2023.13/plantuml-1.2023.13.jar

# Render all diagrams to PNG
plantuml diagrams/*.puml
# or
java -jar plantuml.jar diagrams/*.puml

# Render to SVG
plantuml -tsvg diagrams/*.puml
```

### GitHub
GitHub automatically renders PlantUML diagrams in markdown when you use the following syntax:

```markdown
![Diagram Name](path/to/diagram.puml)
```

## Customization

All diagrams use the `!theme plain` directive for consistency. You can customize:

- **Colors**: Change component colors using `#RRGGBB` hex codes
- **Layout**: Adjust spacing and direction
- **Notes**: Add or modify explanatory notes
- **Styling**: Use skinparams for global styling changes

## Maintenance

When updating the API architecture:
1. Update the corresponding `.puml` file(s)
2. Verify the diagram renders correctly
3. Update the main `README.md` if the flow changes significantly
