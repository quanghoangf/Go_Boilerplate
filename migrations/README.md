# Database Migrations

This directory contains database migration files using [golang-migrate](https://github.com/golang-migrate/migrate).

## Migration Files

Each migration consists of two files:
- `.up.sql` - Contains the forward migration (applying changes)
- `.down.sql` - Contains the reverse migration (rolling back changes)

## File Naming Convention

Migrations follow the pattern: `{version}_{description}.{direction}.sql`

Example:
- `000001_create_users_table.up.sql`
- `000001_create_users_table.down.sql`

## Current Migrations

### 000001_create_users_table
- **Purpose**: Creates the initial users table with proper indexes and constraints
- **Up**: Creates users table, indexes, and update trigger
- **Down**: Drops table, indexes, trigger, and function

## Usage

See the main [README.md](../README.md) for migration commands or use:

```bash
# Apply all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Create new migration
make migrate-create name=your_migration_name

# Check current version
make migrate-version
```

## Best Practices

1. **Test migrations** in development before applying to production
2. **Backup database** before running migrations in production
3. **Make migrations reversible** - always provide proper down migrations
4. **Keep migrations focused** - one logical change per migration
5. **Use transactions** when appropriate (PostgreSQL supports DDL in transactions)
6. **Check for destructive operations** - be careful with data-destroying changes

## Migration States

The golang-migrate tool tracks migration state in a `schema_migrations` table:
- `version` - The migration version number
- `dirty` - Whether the migration failed and needs manual intervention

If a migration fails, the database will be marked as "dirty" and you'll need to:
1. Fix the issue manually
2. Use `migrate force <version>` to mark the database as clean
3. Re-run the migration

## PostgreSQL Specific Notes

- PostgreSQL supports DDL operations in transactions
- Some operations like `CREATE INDEX CONCURRENTLY` cannot run in transactions
- Use `IF EXISTS` and `IF NOT EXISTS` for idempotent operations
- Consider using `CONCURRENTLY` for index creation on large tables in production