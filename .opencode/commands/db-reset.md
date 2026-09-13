---
description: Reset the database to a clean state with fresh migrations and seed data
agent: database
---

Reset the ZarishLog database to a clean state:

1. Drop all tables and schemas
2. Recreate the public schema
3. Run all migrations in order
4. Seed with master data
5. Verify the database is ready

Use `make db-reset` to perform the full reset, or follow these steps manually if needed:

```bash
# Drop and recreate schema
psql -h localhost -U zarishlog -d zarishlog -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# Run migrations
make db-migrate

# Seed data
make db-seed
```

After reset:

- Verify all tables exist
- Check seed data is loaded
- Test a simple query to confirm connectivity
- Update any test fixtures if needed

**Warning**: This will destroy all existing data. Only use in development or when explicitly instructed.
