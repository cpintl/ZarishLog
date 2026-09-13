---
description: Run the complete test suite for all components
agent: build
---

Run the complete test suite for ZarishLog:

1. Run Go API tests with race detection
2. Run business logic package tests
3. Run frontend tests
4. Generate coverage reports
5. Check for any failing tests

Use `make test` to run all tests, or run specific test suites:

```bash
# Go tests
make test-go

# Frontend tests
make test-web

# Integration tests
make test-integration
```

After tests complete:
- Review any failures
- Check code coverage metrics
- Fix any race conditions detected
- Update test fixtures if needed
- Document any flaky tests

If tests fail, provide:
- Failed test names and file locations
- Error messages and stack traces
- Suggested fixes or workarounds