# Backend Domain-Driven Structure Refactoring

**Status**: 📋 Planning
**Date**: 2026-02-09
**Estimated Effort**: 8-10 hours
**Impact**: Major architectural improvement

---

## 📖 Overview

This refactoring transforms the backend from a **flat layer-based structure** to a **domain-driven architecture** with hybrid folders. The goal is to improve code organization, reduce duplication, and prepare for future microservices extraction.

### Current Problems
- ✅ **Functionality conflicts** - Unclear where to add new code
- ✅ **Codebase growing too large** - 45 domain files in flat structure
- ✅ **DTO duplication** - Same DTOs repeated across handlers
- ✅ **Developer friction** - Time wasted finding related code

### Solution
- **Domain-based organization** - Group by business capability, not technical layer
- **Hybrid structure** - Layered folders (handlers, services, repositories, dtos) within each domain
- **DTO consolidation** - Single source of truth per domain
- **Clear boundaries** - Prepare for microservices extraction

---

## 📂 Documentation Structure

This directory contains comprehensive documentation for the refactoring:

### 1. [Domain-Driven Structure Plan](./domain-driven-structure-plan.md) 📘
**What**: Comprehensive plan with full rationale and migration strategy
**When**: Read before starting migration
**Who**: Tech leads, architects, all developers

**Contents**:
- Current state analysis (metrics, issues, duplication)
- Target architecture (structure, domain identification)
- Benefits analysis (7 success criteria)
- Migration plan (6 phases with detailed steps)
- Risk assessment (import cycles, tests, CI/CD)
- Success criteria and validation

**Time to Read**: 30-45 minutes

---

### 2. [Migration Checklist](./migration-checklist.md) ✅
**What**: Step-by-step actionable checklist for migration execution
**When**: Use during migration to track progress
**Who**: Developer executing the migration

**Contents**:
- Phase-by-phase checkboxes
- Validation criteria for each phase
- Git commit messages
- Time tracking per phase
- Issue tracking

**Time to Complete**: 8-10 hours (estimated)

---

### 3. [Quick Reference](./quick-reference.md) 🚀
**What**: Cheat sheet for working with the new structure
**When**: After migration, during daily development
**Who**: All developers (reference guide)

**Contents**:
- Directory structure overview
- Quick commands (testing, building, linting)
- Import path patterns
- Finding code by feature/layer
- Common tasks (adding endpoints, new domains)
- Cross-domain communication patterns
- Common pitfalls and solutions

**Time to Read**: 10-15 minutes

---

## 🎯 Quick Start

### Before Migration

1. **Read the Plan** → [domain-driven-structure-plan.md](./domain-driven-structure-plan.md)
2. **Review with Team** → Discuss concerns, timeline, ownership
3. **Prepare Environment**:
   ```bash
   git checkout -b refactor/domain-driven-structure
   go test ./...  # Ensure all tests pass
   golangci-lint run  # Ensure linter clean
   ```

### During Migration

1. **Follow the Checklist** → [migration-checklist.md](./migration-checklist.md)
2. **Commit After Each Phase** → Incremental, reversible changes
3. **Run Tests Frequently** → Catch issues early
4. **Track Time** → Record actual vs estimated time

### After Migration

1. **Use Quick Reference** → [quick-reference.md](./quick-reference.md)
2. **Update CLAUDE.md** → Document new structure
3. **Share with Team** → Demo new organization
4. **Collect Feedback** → Improve documentation

---

## 📊 Migration Overview

### Structure Transformation

```
❌ BEFORE (Flat)                      ✅ AFTER (Domain-Driven)
backend/internal/                      backend/internal/
├── handlers/          (16 files)     ├── domains/
├── services/          (19 files)     │   ├── auth/           (11 files)
├── repositories/      (10 files)     │   ├── wishlists/      (10 files)
├── middleware/                       │   ├── items/          (5 files)
├── config/                           │   ├── reservations/   (5 files)
├── db/                               │   ├── storage/        (3 files)
├── cache/                            │   └── health/         (2 files)
├── auth/                             └── shared/
├── encryption/                           ├── middleware/
├── validation/                           ├── config/
└── analytics/                            ├── db/
                                          ├── cache/
                                          ├── encryption/
                                          ├── validation/
                                          └── analytics/
```

### 6 Phases (8-10 hours)

| Phase | Description | Time | Risk |
|-------|-------------|------|------|
| **0** | Preparation - checkpoints, baseline | 1-2h | Low |
| **1** | Create domain structure | 30m | Low |
| **2** | Move cross-cutting concerns to shared/ | 1h | Low |
| **3** | Migrate domains one-by-one | 3-4h | Medium |
| **4** | Consolidate DTOs | 2h | Low |
| **5** | Update all imports | 1h | Low |
| **6** | Validation & cleanup | 1h | Low |

### Domains

| Domain | Files | Responsibility |
|--------|-------|---------------|
| **auth** | 11 | Authentication, OAuth, User management |
| **wishlists** | 10 | Wishlist CRUD, templates, relationships |
| **items** | 5 | Gift items as independent resources |
| **reservations** | 5 | Reservations, purchases, notifications |
| **storage** | 3 | S3 file uploads, media |
| **health** | 2 | System monitoring, health checks |

---

## ✅ Success Criteria

### Quantitative
- [ ] All 28 tests passing (100%)
- [ ] Zero import cycles
- [ ] Build succeeds without errors
- [ ] Test coverage ≥ baseline
- [ ] Linter clean (0 new issues)

### Qualitative
- [ ] Each domain is self-contained
- [ ] New developers can navigate easily
- [ ] Microservices-ready architecture
- [ ] Clear "where does this go?" answers

---

## 🚨 Key Risks & Mitigation

### Import Cycles
**Risk**: Circular dependencies between domains
**Mitigation**: Keep domains independent, use interfaces, shared/ for common types

### Test Breakage
**Risk**: 28 test files with hardcoded imports
**Mitigation**: Update one domain at a time, run tests after each phase

### CI/CD Pipeline
**Risk**: Path-specific configurations
**Mitigation**: Review and update CI configs before migration

---

## 📋 TODO After Migration

### Immediate (Within 1 day)
- [ ] Update `CLAUDE.md` with new structure
- [ ] Update main `README.md` architecture section
- [ ] Notify team about structure change
- [ ] Create PR with detailed summary

### Short-term (Within 1 week)
- [ ] Update onboarding documentation
- [ ] Update code review guidelines
- [ ] Share IDE workspace settings
- [ ] Monitor for production issues

### Long-term (Within 1 month)
- [ ] Conduct team training on domain structure
- [ ] Document domain design patterns
- [ ] Create `scripts/create-domain.sh` generator
- [ ] Track "time to find code" improvements

---

## 🤝 Team Collaboration

### Roles

| Role | Responsibility |
|------|---------------|
| **Migration Lead** | Execute migration, follow checklist |
| **Code Reviewers** | Verify domain boundaries, DTO consolidation, test coverage |
| **Tech Lead** | Final approval, architectural decisions |
| **Team** | Provide feedback, adopt new structure |

### Communication Plan

**Before**: Share plan, gather feedback, schedule migration
**During**: Update team after each phase, escalate blockers
**After**: Demo new structure, document lessons learned

---

## 📚 Resources

### Internal Documentation
- [Go Architecture Guide](../Go-Architecture-Guide.md) - 3-layer architecture principles
- [Cross-Domain Architecture Plan](../plans/00-cross-domain-architecture-plan.md)
- [CLAUDE.md](../../CLAUDE.md) - Project conventions

### External Resources
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html) by Martin Fowler
- [Go Project Layout](https://github.com/golang-standards/project-layout) - Standard Go structure
- [Package Oriented Design](https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html) by Arden Labs

---

## 📞 Support

### Questions?
- Review [Quick Reference](./quick-reference.md) for common patterns
- Check [Comprehensive Plan](./domain-driven-structure-plan.md) for detailed explanations
- Ask in team chat or create discussion issue

### Issues During Migration?
- Check [Migration Checklist](./migration-checklist.md) validation steps
- Review [Risk Assessment](./domain-driven-structure-plan.md#-risk-assessment) section
- Rollback to last checkpoint if needed

---

## 📝 Lessons Learned (To Be Filled)

### What Went Well
- _To be filled after migration_

### What Was Challenging
- _To be filled after migration_

### What Would We Do Differently
- _To be filled after migration_

### Time Estimate Accuracy
- **Estimated**: 8-10 hours
- **Actual**: _____ hours
- **Variance**: _____ hours

---

**Created**: 2026-02-09
**Status**: Planning → In Progress → Completed
**Next Steps**: Read comprehensive plan, review with team, schedule migration

---

## 🎉 Expected Benefits

After successful migration, expect:

- **70% reduction** in "file hopping" when working on features
- **Faster onboarding** - new developers find code intuitively
- **Microservices-ready** - clean boundaries for future extraction
- **Reduced duplication** - single source of truth for DTOs
- **Clear ownership** - one team per domain
- **Sustainable growth** - structure scales to 100+ files

---

**Let's make the codebase more maintainable! 🚀**
