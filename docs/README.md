# Project Documentation

## Structure

```
docs/
├── README.md              # This file
├── archive/               # Historical documents (completed tasks, old plans)
├── active/                # Current working documents (in-progress tasks)
├── architecture/          # Architecture documentation, guides, and decisions
├── plans/                 # Implementation plans (cross-domain, security, mobile, backend)
└── refactoring/           # Refactoring documentation (if active)
```

## Directories

### 📁 archive/
Completed implementation summaries, old review reports, and historical documents.
- PHASE6_IMPLEMENTATION_SUMMARY.md
- T087_IMPLEMENTATION_SUMMARY.md
- T088_T089_IMPLEMENTATION_SUMMARY.md
- CHANGES.md (sqlc→sqlx migration history)

### 📁 active/
Current work-in-progress documents and task trackers.
- pr-issues-todo.md - GitHub PR issues tracker (85 open issues)
- mobile-app-completion-plan.md - Mobile app completion roadmap

### 📁 architecture/
Core architecture documentation, security audits, and technical guides.
- Go-Architecture-Guide.md - Backend architecture patterns
- DEEP_LINKING.md - Mobile deep linking implementation
- NAVIGATION_ARCHITECTURE.md - Cross-platform navigation
- MOBILE_REDIRECTION.md - Web-to-mobile redirection system
- SECURITY_AUDIT.md - Security audit and compliance
- PERFORMANCE.md - Performance optimization guide
- API.md - API documentation and usage guide

### 📁 plans/
Implementation plans for different aspects of the project.
- 00-cross-domain-architecture-plan.md - Auth, CORS, handoff flow
- 01-frontend-security-and-quality-plan.md - Frontend security
- 02-mobile-app-completion-plan.md - Mobile app features
- 03-api-backend-improvements-plan.md - Backend improvements

### 📁 refactoring/
Active refactoring documentation (Domain-Driven Structure migration).

## Usage Guidelines

- **archive/**: Don't modify - reference only
- **active/**: Update regularly as work progresses
- **architecture/**: Update when architectural decisions change
- **plans/**: Mark complete as phases finish

## Last Updated
2026-02-16
