---
description: Lint all code across the entire monorepo
agent: build
---

Lint all code in the ZarishLog monorepo:

1. Run Go linter (golangci-lint or go vet)
2. Run frontend linter (ESLint)
3. Check for formatting issues
4. Validate code style consistency
5. Report any violations

Use `make lint` to lint everything, or run specific linters:

```bash
# Go linting
make lint-go

# Frontend linting
make lint-web
```

After linting:
- Fix any critical violations
- Address warnings that could become errors
- Ensure consistent code style across the monorepo
- Update linting rules if new patterns emerge

Common issues to watch for:
- Unused imports or variables
- Missing error handling
- Inconsistent naming conventions
- Security vulnerabilities
- Performance anti-patterns