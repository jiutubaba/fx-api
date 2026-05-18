# Migration From newAPI

## Tool

Active migration tool:

```text
backend/cmd/migrate-newapi
```

Parameters:

- `--source-dsn`
- `--target-dsn`
- `--source-name=prod|staging`
- `--mode=dry-run|apply`
- `--report-dir`
- `--legal-doc`

## Current Rules

- Old raw tables are imported into `legacy_newapi` for archive.
- Compatible core data is mapped into sub2api tables.
- Password bcrypt hashes are preserved.
- Old quota values are converted using `QuotaPerUnit`.
- Old account-pool accounts are migrated into `accounts`.
- Old `AutoGroups` bindings are expanded into the new `auto` group.
- Empty active OpenAI key groups receive fallback active schedulable OpenAI accounts where migration rules require continuity.

## Known Warnings

- Some old consumption logs may be archive-only when old token/channel rows cannot map to target rows.
- Legacy custom OAuth bindings are archive-only unless compatibility is explicitly verified.
- TOTP/Passkey credentials require compatibility verification before active migration.

