---
description: Build and deploy the application to production
agent: devops
---

Build and deploy ZarishLog to production:

1. Run all tests to ensure stability
2. Build Go binary with production flags
3. Build frontend for production
4. Build Docker images with version tag
5. Push images to container registry
6. Create release tag if version bump needed

Use `make publish` for the full deployment pipeline, or `make release` to create a versioned release.

Before deploying:
- Ensure all tests pass (`make test`)
- Lint all code (`make lint`)
- Validate configuration (`make validate`)
- Check for any uncommitted changes

After deploying:
- Verify services are running
- Check application health endpoints
- Monitor logs for any issues
- Update docs/STATUS.md if significant changes were made