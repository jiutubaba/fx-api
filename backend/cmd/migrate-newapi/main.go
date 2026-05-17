package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jiutubaba/fx-api/internal/repository"

	_ "github.com/lib/pq"
)

const (
	modeDryRun = "dry-run"
	modeApply  = "apply"

	sourceProd    = "prod"
	sourceStaging = "staging"

	defaultQuotaPerUnit = 500000.0
)

type options struct {
	sourceDSN  string
	targetDSN  string
	sourceName string
	mode       string
	reportDir  string
	legalDoc   string
}

type report struct {
	SourceName        string            `json:"source_name"`
	Mode              string            `json:"mode"`
	StartedAt         time.Time         `json:"started_at"`
	FinishedAt        time.Time         `json:"finished_at"`
	QuotaPerUnit      float64           `json:"quota_per_unit"`
	MigrationsApplied bool              `json:"migrations_applied"`
	LegacyArchive     []tableReport     `json:"legacy_archive"`
	Transforms        []transformReport `json:"transforms"`
	Summary           map[string]int64  `json:"summary"`
	Warnings          []string          `json:"warnings"`
}

type tableReport struct {
	Table        string `json:"table"`
	SourceRows   int64  `json:"source_rows"`
	ArchivedRows int64  `json:"archived_rows,omitempty"`
	Status       string `json:"status"`
}

type transformReport struct {
	Name         string `json:"name"`
	RowsAffected int64  `json:"rows_affected"`
	Skipped      int64  `json:"skipped,omitempty"`
}

type sourceColumn struct {
	Name string
	Type string
}

func main() {
	var opts options
	flag.StringVar(&opts.sourceDSN, "source-dsn", "", "new-api source PostgreSQL DSN")
	flag.StringVar(&opts.targetDSN, "target-dsn", "", "fuxi-api target PostgreSQL DSN")
	flag.StringVar(&opts.sourceName, "source-name", "", "source label: prod or staging")
	flag.StringVar(&opts.mode, "mode", modeDryRun, "migration mode: dry-run or apply")
	flag.StringVar(&opts.reportDir, "report-dir", "migration-reports", "directory for JSON and markdown reports")
	flag.StringVar(&opts.legalDoc, "legal-doc", "", "optional path to Fuxi API legal markdown")
	flag.Parse()

	if err := validateOptions(opts); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := run(ctx, opts); err != nil {
		log.Fatal(err)
	}
}

func validateOptions(opts options) error {
	if strings.TrimSpace(opts.sourceDSN) == "" {
		return errors.New("--source-dsn is required")
	}
	if strings.TrimSpace(opts.targetDSN) == "" {
		return errors.New("--target-dsn is required")
	}
	switch opts.sourceName {
	case sourceProd, sourceStaging:
	default:
		return fmt.Errorf("--source-name must be %q or %q", sourceProd, sourceStaging)
	}
	switch opts.mode {
	case modeDryRun, modeApply:
	default:
		return fmt.Errorf("--mode must be %q or %q", modeDryRun, modeApply)
	}
	if strings.TrimSpace(opts.reportDir) == "" {
		return errors.New("--report-dir is required")
	}
	return nil
}

func run(ctx context.Context, opts options) error {
	sourceDB, err := sql.Open("postgres", opts.sourceDSN)
	if err != nil {
		return fmt.Errorf("open source db: %w", err)
	}
	defer func() { _ = sourceDB.Close() }()
	targetDB, err := sql.Open("postgres", opts.targetDSN)
	if err != nil {
		return fmt.Errorf("open target db: %w", err)
	}
	defer func() { _ = targetDB.Close() }()

	if err := sourceDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping source db: %w", err)
	}
	if err := targetDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping target db: %w", err)
	}

	rep := &report{
		SourceName:   opts.sourceName,
		Mode:         opts.mode,
		StartedAt:    time.Now(),
		Summary:      map[string]int64{},
		QuotaPerUnit: defaultQuotaPerUnit,
	}

	tables, err := listSourceTables(ctx, sourceDB)
	if err != nil {
		return fmt.Errorf("list source tables: %w", err)
	}
	sort.Strings(tables)
	for _, table := range tables {
		count, err := countRows(ctx, sourceDB, "public", table)
		if err != nil {
			return fmt.Errorf("count source table %s: %w", table, err)
		}
		rep.LegacyArchive = append(rep.LegacyArchive, tableReport{
			Table:      table,
			SourceRows: count,
			Status:     "planned",
		})
	}

	qpu, err := detectSourceQuotaPerUnit(ctx, sourceDB)
	if err != nil {
		rep.Warnings = append(rep.Warnings, "failed to read QuotaPerUnit from source options; using 500000")
	} else {
		rep.QuotaPerUnit = qpu
	}

	if opts.mode == modeApply {
		if err := repository.ApplyMigrations(ctx, targetDB); err != nil {
			return fmt.Errorf("apply target migrations: %w", err)
		}
		rep.MigrationsApplied = true

		archiveReports, err := archiveLegacyTables(ctx, sourceDB, targetDB, tables)
		if err != nil {
			return err
		}
		rep.LegacyArchive = archiveReports

		if err := installHelperFunctions(ctx, targetDB); err != nil {
			return fmt.Errorf("install helper functions: %w", err)
		}
		defer func() {
			_, _ = targetDB.ExecContext(context.Background(), "DROP FUNCTION IF EXISTS public.fuxi_migrate_safe_jsonb(text, jsonb)")
		}()

		transforms, warnings, err := transformCoreData(ctx, targetDB, opts)
		if err != nil {
			return err
		}
		rep.Transforms = append(rep.Transforms, transforms...)
		rep.Warnings = append(rep.Warnings, warnings...)
		if err := resetTargetSequences(ctx, targetDB); err != nil {
			return fmt.Errorf("reset target sequences: %w", err)
		}
	}

	if err := fillSummary(ctx, targetDB, rep); err != nil {
		rep.Warnings = append(rep.Warnings, "failed to collect target summary: "+err.Error())
	}
	rep.FinishedAt = time.Now()

	if err := writeReports(opts.reportDir, rep); err != nil {
		return fmt.Errorf("write reports: %w", err)
	}
	return nil
}

func listSourceTables(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
SELECT c.relname
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public'
  AND c.relkind IN ('r', 'p')
ORDER BY c.relname`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, rows.Err()
}

func listSourceColumns(ctx context.Context, db *sql.DB, table string) ([]sourceColumn, error) {
	rows, err := db.QueryContext(ctx, `
SELECT a.attname, pg_catalog.format_type(a.atttypid, a.atttypmod)
FROM pg_attribute a
JOIN pg_class c ON c.oid = a.attrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public'
  AND c.relname = $1
  AND a.attnum > 0
  AND NOT a.attisdropped
ORDER BY a.attnum`, table)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var cols []sourceColumn
	for rows.Next() {
		var col sourceColumn
		if err := rows.Scan(&col.Name, &col.Type); err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}
	return cols, rows.Err()
}

func countRows(ctx context.Context, db *sql.DB, schema, table string) (int64, error) {
	var count int64
	err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", quoteIdent(schema), quoteIdent(table))).Scan(&count)
	return count, err
}

func detectSourceQuotaPerUnit(ctx context.Context, db *sql.DB) (float64, error) {
	exists, err := sourceTableExists(ctx, db, "options")
	if err != nil {
		return defaultQuotaPerUnit, err
	}
	if !exists {
		return defaultQuotaPerUnit, nil
	}
	var raw sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT value FROM public.options WHERE key = 'QuotaPerUnit' LIMIT 1`).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultQuotaPerUnit, nil
		}
		return defaultQuotaPerUnit, err
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(raw.String), 64)
	if err != nil || value <= 0 {
		return defaultQuotaPerUnit, nil
	}
	return value, nil
}

func sourceTableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM information_schema.tables
  WHERE table_schema = 'public' AND table_name = $1
)`, table).Scan(&exists)
	return exists, err
}

func targetTableExists(ctx context.Context, db *sql.DB, schema, table string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM information_schema.tables
  WHERE table_schema = $1 AND table_name = $2
)`, schema, table).Scan(&exists)
	return exists, err
}

func archiveLegacyTables(ctx context.Context, sourceDB, targetDB *sql.DB, tables []string) ([]tableReport, error) {
	if _, err := targetDB.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS legacy_newapi"); err != nil {
		return nil, fmt.Errorf("create legacy_newapi schema: %w", err)
	}

	reports := make([]tableReport, 0, len(tables))
	for _, table := range tables {
		sourceCount, err := countRows(ctx, sourceDB, "public", table)
		if err != nil {
			return nil, fmt.Errorf("count source table %s: %w", table, err)
		}
		cols, err := listSourceColumns(ctx, sourceDB, table)
		if err != nil {
			return nil, fmt.Errorf("list columns for source table %s: %w", table, err)
		}
		if len(cols) == 0 {
			reports = append(reports, tableReport{Table: table, SourceRows: sourceCount, Status: "skipped_empty_schema"})
			continue
		}
		if err := recreateLegacyTable(ctx, targetDB, table, cols); err != nil {
			return nil, err
		}
		archived, err := copyTableData(ctx, sourceDB, targetDB, table, cols)
		if err != nil {
			return nil, err
		}
		reports = append(reports, tableReport{
			Table:        table,
			SourceRows:   sourceCount,
			ArchivedRows: archived,
			Status:       "archived",
		})
	}
	return reports, nil
}

func recreateLegacyTable(ctx context.Context, db *sql.DB, table string, cols []sourceColumn) error {
	colDefs := make([]string, 0, len(cols))
	for _, col := range cols {
		colDefs = append(colDefs, fmt.Sprintf("%s %s", quoteIdent(col.Name), sanitizeType(col.Type)))
	}
	fullName := "legacy_newapi." + quoteIdent(table)
	stmts := []string{
		"DROP TABLE IF EXISTS " + fullName + " CASCADE",
		"CREATE TABLE " + fullName + " (" + strings.Join(colDefs, ", ") + ")",
		fmt.Sprintf("COMMENT ON TABLE %s IS %s", fullName, quoteLiteral("Read-only legacy new-api archive imported by migrate-newapi")),
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("prepare legacy archive table %s: %w", table, err)
		}
	}
	return nil
}

func sanitizeType(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "text"
	}
	allowed := regexp.MustCompile(`^[a-zA-Z0-9_\[\]\s\(\),\."]+$`)
	if !allowed.MatchString(raw) {
		return "text"
	}
	return raw
}

func copyTableData(ctx context.Context, sourceDB, targetDB *sql.DB, table string, cols []sourceColumn) (int64, error) {
	colNames := make([]string, 0, len(cols))
	for _, col := range cols {
		colNames = append(colNames, quoteIdent(col.Name))
	}
	selectSQL := fmt.Sprintf("SELECT %s FROM public.%s", strings.Join(colNames, ", "), quoteIdent(table))
	rows, err := sourceDB.QueryContext(ctx, selectSQL)
	if err != nil {
		return 0, fmt.Errorf("read source table %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()

	tx, err := targetDB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin archive tx for %s: %w", table, err)
	}
	defer func() { _ = tx.Rollback() }()

	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	insertSQL := fmt.Sprintf(
		"INSERT INTO legacy_newapi.%s (%s) VALUES (%s)",
		quoteIdent(table),
		strings.Join(colNames, ", "),
		strings.Join(placeholders, ", "),
	)
	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return 0, fmt.Errorf("prepare archive insert for %s: %w", table, err)
	}
	defer func() { _ = stmt.Close() }()

	raw := make([]sql.RawBytes, len(cols))
	scanArgs := make([]any, len(cols))
	values := make([]any, len(cols))
	for i := range raw {
		scanArgs[i] = &raw[i]
	}

	var copied int64
	for rows.Next() {
		for i := range raw {
			raw[i] = nil
			values[i] = nil
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return 0, fmt.Errorf("scan source row for %s: %w", table, err)
		}
		for i := range raw {
			if raw[i] != nil {
				values[i] = string(raw[i])
			}
		}
		if _, err := stmt.ExecContext(ctx, values...); err != nil {
			return 0, fmt.Errorf("insert legacy row into %s: %w", table, err)
		}
		copied++
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate source rows for %s: %w", table, err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit archive tx for %s: %w", table, err)
	}
	return copied, nil
}

func installHelperFunctions(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE OR REPLACE FUNCTION public.fuxi_migrate_safe_jsonb(input text, fallback jsonb DEFAULT '{}'::jsonb)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
BEGIN
  IF input IS NULL OR btrim(input) = '' THEN
    RETURN fallback;
  END IF;
  RETURN input::jsonb;
EXCEPTION WHEN others THEN
  RETURN fallback;
END;
$$`)
	return err
}

func transformCoreData(ctx context.Context, db *sql.DB, opts options) ([]transformReport, []string, error) {
	var reports []transformReport
	var warnings []string

	steps := []struct {
		name      string
		required  []string
		statement string
	}{
		{"groups", []string{"users", "tokens", "channels"}, migrateGroupsSQL},
		{"users", []string{"users"}, migrateUsersSQL},
		{"auth_identities", []string{"users"}, migrateAuthIdentitiesSQL},
		{"api_keys", []string{"tokens"}, migrateAPIKeysSQL},
		{"accounts", []string{"channels"}, migrateAccountsSQL},
		{"account_groups", []string{"channels"}, migrateAccountGroupsSQL},
		{"redeem_codes", []string{"redemptions"}, migrateRedeemCodesSQL},
		{"usage_logs", []string{"logs"}, migrateUsageLogsSQL},
		{"payment_orders", []string{"top_ups"}, migratePaymentOrdersSQL},
	}

	for _, step := range steps {
		ok, missing, err := legacyTablesExist(ctx, db, step.required)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			warnings = append(warnings, fmt.Sprintf("skip %s: missing legacy table(s): %s", step.name, strings.Join(missing, ", ")))
			continue
		}
		rows, err := execRowsAffected(ctx, db, step.statement)
		if err != nil {
			return nil, nil, fmt.Errorf("migrate %s: %w", step.name, err)
		}
		reports = append(reports, transformReport{Name: step.name, RowsAffected: rows})
	}

	settingsRows, err := migrateSettings(ctx, db, opts)
	if err != nil {
		return nil, nil, err
	}
	reports = append(reports, transformReport{Name: "settings", RowsAffected: settingsRows})

	skippedLogs, err := countSkippedUsageLogs(ctx, db)
	if err == nil && skippedLogs > 0 {
		warnings = append(warnings, fmt.Sprintf("%d consumption logs were archived only because token_id/channel_id did not map to target rows", skippedLogs))
	}
	if n, err := countRowsWhere(ctx, db, "legacy_newapi", "users", "COALESCE(github_id, '') = '' AND COALESCE(oidc_id, '') = '' AND COALESCE(wechat_id, '') = '' AND COALESCE(linux_do_id, '') = ''"); err == nil {
		reports = append(reports, transformReport{Name: "users_without_supported_oauth_identity", RowsAffected: n})
	}
	if exists, _ := targetTableExists(ctx, db, "legacy_newapi", "passkeys"); exists {
		warnings = append(warnings, "legacy passkeys were archived only; passkey credential compatibility must be verified before enabling migration")
	}
	if exists, _ := targetTableExists(ctx, db, "legacy_newapi", "user_oauth_bindings"); exists {
		warnings = append(warnings, "legacy custom OAuth bindings were archived only; provider-specific compatibility must be verified before enabling migration")
	}
	return reports, warnings, nil
}

func legacyTablesExist(ctx context.Context, db *sql.DB, tables []string) (bool, []string, error) {
	var missing []string
	for _, table := range tables {
		exists, err := targetTableExists(ctx, db, "legacy_newapi", table)
		if err != nil {
			return false, nil, err
		}
		if !exists {
			missing = append(missing, table)
		}
	}
	return len(missing) == 0, missing, nil
}

func execRowsAffected(ctx context.Context, db *sql.DB, stmt string) (int64, error) {
	res, err := db.ExecContext(ctx, stmt)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, nil
	}
	return rows, nil
}

func migrateSettings(ctx context.Context, db *sql.DB, opts options) (int64, error) {
	var rows int64
	if ok, _, err := legacyTablesExist(ctx, db, []string{"options"}); err != nil {
		return 0, err
	} else if ok {
		n, err := execRowsAffected(ctx, db, `
INSERT INTO settings (key, value, updated_at)
SELECT LEFT(key, 100), value, NOW()
FROM legacy_newapi.options
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`)
		if err != nil {
			return 0, fmt.Errorf("copy legacy options to settings: %w", err)
		}
		rows += n
	}

	siteName := "伏羲API"
	publicURL := "https://fuxiapi.top"
	registrationEnabled := "true"
	if opts.sourceName == sourceStaging {
		siteName = "伏羲API-预发"
		publicURL = "https://staging.fuxiapi.top"
		registrationEnabled = "false"
	}

	overrides := map[string]string{
		"site_name":                  siteName,
		"site_subtitle":              "AI API 网关与额度管理平台",
		"api_base_url":               publicURL,
		"frontend_url":               publicURL,
		"registration_enabled":       registrationEnabled,
		"login_agreement_enabled":    "true",
		"login_agreement_mode":       "modal",
		"login_agreement_updated_at": time.Now().Format("2006-01-02"),
	}
	if logo, err := getLegacyOption(ctx, db, "Logo"); err == nil && strings.TrimSpace(logo) != "" {
		overrides["site_logo"] = logo
	}
	if docs := loadLoginAgreementDocuments(opts.legalDoc); len(docs) > 0 {
		raw, err := json.Marshal(docs)
		if err != nil {
			return 0, fmt.Errorf("marshal login agreement docs: %w", err)
		}
		overrides["login_agreement_documents"] = string(raw)
	}

	for key, value := range mappedSettingOverrides(ctx, db) {
		if _, exists := overrides[key]; !exists {
			overrides[key] = value
		}
	}
	for key, value := range overrides {
		n, err := upsertSetting(ctx, db, key, value)
		if err != nil {
			return rows, err
		}
		rows += n
	}
	return rows, nil
}

func mappedSettingOverrides(ctx context.Context, db *sql.DB) map[string]string {
	pairs := map[string]string{
		"RegisterEnabled":          "registration_enabled",
		"EmailVerificationEnabled": "email_verify_enabled",
		"GitHubOAuthEnabled":       "github_oauth_enabled",
		"GitHubClientId":           "github_oauth_client_id",
		"GitHubClientSecret":       "github_oauth_client_secret",
		"LinuxDOOAuthEnabled":      "linuxdo_connect_enabled",
		"WeChatAuthEnabled":        "wechat_connect_enabled",
		"TurnstileCheckEnabled":    "turnstile_enabled",
		"TurnstileSiteKey":         "turnstile_site_key",
		"TurnstileSecretKey":       "turnstile_secret_key",
		"TopUpLink":                "purchase_subscription_url",
		"HomePageContent":          "home_content",
		"About":                    "contact_info",
	}
	result := make(map[string]string)
	for oldKey, newKey := range pairs {
		value, err := getLegacyOption(ctx, db, oldKey)
		if err == nil && strings.TrimSpace(value) != "" {
			result[newKey] = normalizeBoolSetting(value)
		}
	}
	if url, err := getLegacyOption(ctx, db, "ServerAddress"); err == nil && strings.TrimSpace(url) != "" {
		result["api_base_url"] = strings.TrimSpace(url)
		result["frontend_url"] = strings.TrimSpace(url)
	}
	return result
}

func normalizeBoolSetting(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "false":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return value
	}
}

func getLegacyOption(ctx context.Context, db *sql.DB, key string) (string, error) {
	var value sql.NullString
	err := db.QueryRowContext(ctx, `SELECT value FROM legacy_newapi.options WHERE key = $1 LIMIT 1`, key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value.String, nil
}

func upsertSetting(ctx context.Context, db *sql.DB, key, value string) (int64, error) {
	res, err := db.ExecContext(ctx, `
INSERT INTO settings (key, value, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`, key, value)
	if err != nil {
		return 0, fmt.Errorf("upsert setting %s: %w", key, err)
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

type loginAgreementDocument struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	ContentMD string `json:"content_md"`
}

func loadLoginAgreementDocuments(path string) []loginAgreementDocument {
	candidates := []string{}
	if strings.TrimSpace(path) != "" {
		candidates = append(candidates, path)
	}
	candidates = append(candidates,
		filepath.Join("deploy", "fuxi", "legal", "user-agreement.md"),
		filepath.Join("..", "deploy", "fuxi", "legal", "user-agreement.md"),
	)
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err == nil && len(strings.TrimSpace(string(data))) > 0 {
			return []loginAgreementDocument{{
				ID:        "terms",
				Title:     "伏羲API用户协议",
				ContentMD: string(data),
			}}
		}
	}
	return []loginAgreementDocument{{
		ID:        "terms",
		Title:     "伏羲API用户协议",
		ContentMD: "请遵守伏羲API服务条款、平台规则、上游服务条款和适用法律法规。",
	}}
}

func resetTargetSequences(ctx context.Context, db *sql.DB) error {
	tables := []string{"users", "api_keys", "accounts", "redeem_codes", "payment_orders"}
	for _, table := range tables {
		exists, err := targetTableExists(ctx, db, "public", table)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		stmt := fmt.Sprintf(`
SELECT setval(pg_get_serial_sequence(%s, 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM %s), 1), 1), true)`,
			quoteLiteral(table), quoteIdent(table))
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("reset sequence for %s: %w", table, err)
		}
	}
	return nil
}

func countSkippedUsageLogs(ctx context.Context, db *sql.DB) (int64, error) {
	ok, _, err := legacyTablesExist(ctx, db, []string{"logs"})
	if err != nil || !ok {
		return 0, err
	}
	var count int64
	err = db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM legacy_newapi.logs l
LEFT JOIN api_keys ak ON ak.id = l.token_id
LEFT JOIN accounts a ON a.id = l.channel_id
WHERE l.type = 2 AND (ak.id IS NULL OR a.id IS NULL)`).Scan(&count)
	return count, err
}

func countRowsWhere(ctx context.Context, db *sql.DB, schema, table, where string) (int64, error) {
	var count int64
	stmt := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s WHERE %s", quoteIdent(schema), quoteIdent(table), where)
	err := db.QueryRowContext(ctx, stmt).Scan(&count)
	return count, err
}

func fillSummary(ctx context.Context, db *sql.DB, rep *report) error {
	for _, table := range []string{"users", "api_keys", "accounts", "groups", "account_groups", "usage_logs", "redeem_codes", "payment_orders"} {
		exists, err := targetTableExists(ctx, db, "public", table)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		count, err := countRows(ctx, db, "public", table)
		if err != nil {
			return err
		}
		rep.Summary[table] = count
	}
	if exists, err := targetTableExists(ctx, db, "public", "users"); err == nil && exists {
		var cents sql.NullFloat64
		if err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(balance), 0) FROM users WHERE deleted_at IS NULL`).Scan(&cents); err == nil {
			rep.Summary["user_balance_cents_x100000000"] = int64(cents.Float64 * 100000000)
		}
	}
	return nil
}

func writeReports(reportDir string, rep *report) error {
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		return err
	}
	stamp := time.Now().Format("20060102-150405")
	base := filepath.Join(reportDir, fmt.Sprintf("%s-%s", rep.SourceName, stamp))
	jsonPath := base + ".json"
	mdPath := base + ".md"
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(mdPath, []byte(renderMarkdownReport(rep)), 0o644); err != nil {
		return err
	}
	log.Printf("migration reports written: %s, %s", jsonPath, mdPath)
	return nil
}

func renderMarkdownReport(rep *report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# newAPI -> 伏羲API migration report\n\n")
	fmt.Fprintf(&b, "- Source: `%s`\n", rep.SourceName)
	fmt.Fprintf(&b, "- Mode: `%s`\n", rep.Mode)
	fmt.Fprintf(&b, "- Started: `%s`\n", rep.StartedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "- Finished: `%s`\n", rep.FinishedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "- QuotaPerUnit: `%g`\n", rep.QuotaPerUnit)
	fmt.Fprintf(&b, "- Target migrations applied: `%t`\n\n", rep.MigrationsApplied)

	_, _ = b.WriteString("## Legacy Archive\n\n")
	_, _ = b.WriteString("| Table | Source Rows | Archived Rows | Status |\n")
	_, _ = b.WriteString("|---|---:|---:|---|\n")
	for _, item := range rep.LegacyArchive {
		fmt.Fprintf(&b, "| `%s` | %d | %d | %s |\n", item.Table, item.SourceRows, item.ArchivedRows, item.Status)
	}

	_, _ = b.WriteString("\n## Transforms\n\n")
	_, _ = b.WriteString("| Step | Rows Affected | Skipped |\n")
	_, _ = b.WriteString("|---|---:|---:|\n")
	for _, item := range rep.Transforms {
		fmt.Fprintf(&b, "| `%s` | %d | %d |\n", item.Name, item.RowsAffected, item.Skipped)
	}

	if len(rep.Summary) > 0 {
		keys := make([]string, 0, len(rep.Summary))
		for key := range rep.Summary {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		_, _ = b.WriteString("\n## Target Summary\n\n")
		for _, key := range keys {
			fmt.Fprintf(&b, "- `%s`: %d\n", key, rep.Summary[key])
		}
	}

	if len(rep.Warnings) > 0 {
		_, _ = b.WriteString("\n## Warnings\n\n")
		for _, warning := range rep.Warnings {
			fmt.Fprintf(&b, "- %s\n", warning)
		}
	}
	return b.String()
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func quoteLiteral(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}

const quotaPerUnitCTE = `
qpu AS (
  SELECT COALESCE(
    (
      SELECT CASE
        WHEN value ~ '^[0-9]+(\.[0-9]+)?$' THEN value::double precision
        ELSE NULL
      END
      FROM legacy_newapi.options
      WHERE key = 'QuotaPerUnit'
      LIMIT 1
    ),
    500000.0
  ) AS value
)`

const migrateGroupsSQL = `
WITH raw_names AS (
  SELECT unnest(string_to_array(COALESCE(NULLIF("group", ''), 'default'), ',')) AS name FROM legacy_newapi.users
  UNION ALL
  SELECT unnest(string_to_array(COALESCE(NULLIF("group", ''), 'default'), ',')) AS name FROM legacy_newapi.tokens
  UNION ALL
  SELECT unnest(string_to_array(COALESCE(NULLIF("group", ''), 'default'), ',')) AS name FROM legacy_newapi.channels
),
names AS (
  SELECT DISTINCT NULLIF(btrim(name), '') AS name
  FROM raw_names
)
INSERT INTO groups (name, description, rate_multiplier, is_exclusive, status, platform, subscription_type, created_at, updated_at)
SELECT name, 'Migrated from new-api group ' || name, 1.0, false, 'active', 'anthropic', 'standard', NOW(), NOW()
FROM names
WHERE name IS NOT NULL
ON CONFLICT (name) WHERE deleted_at IS NULL
DO UPDATE SET updated_at = EXCLUDED.updated_at`

const migrateUsersSQL = `
WITH ` + quotaPerUnitCTE + `,
base AS (
  SELECT
    u.*,
    CASE
      WHEN btrim(COALESCE(u.email, '')) <> '' THEN lower(btrim(u.email))
      WHEN btrim(COALESCE(u.username, '')) <> '' THEN lower(regexp_replace(btrim(u.username), '[^a-zA-Z0-9._%+-]+', '_', 'g')) || '@legacy.fuxi.local'
      ELSE 'legacy-user-' || u.id::text || '@legacy.fuxi.local'
    END AS email_base
  FROM legacy_newapi.users u
),
dedup AS (
  SELECT
    base.*,
    CASE
      WHEN COUNT(*) OVER (PARTITION BY email_base) > 1
      THEN regexp_replace(email_base, '@', '+' || id::text || '@')
      ELSE email_base
    END AS target_email
  FROM base
)
INSERT INTO users (
  id, email, password_hash, role, balance, concurrency, status, username, notes,
  created_at, updated_at, deleted_at, last_login_at, total_recharged
)
SELECT
  id::bigint,
  LEFT(target_email, 255),
  COALESCE(NULLIF(password, ''), 'legacy-password-not-set'),
  CASE WHEN role >= 10 THEN 'admin' ELSE 'user' END,
  (COALESCE(quota, 0)::double precision / qpu.value),
  5,
  CASE WHEN status = 1 THEN 'active' ELSE 'disabled' END,
  LEFT(COALESCE(NULLIF(username, ''), display_name, 'legacy-user-' || id::text), 100),
  concat_ws(E'\n',
    NULLIF(remark, ''),
    CASE WHEN COALESCE(email, '') = '' THEN 'legacy_newapi: email was empty; synthetic email assigned' ELSE NULL END,
    CASE WHEN COALESCE(setting, '') <> '' THEN 'legacy_newapi.setting=' || setting ELSE NULL END
  ),
  CASE WHEN COALESCE(created_at, 0) > 0 THEN to_timestamp(created_at) ELSE NOW() END,
  NOW(),
  deleted_at,
  CASE WHEN COALESCE(last_login_at, 0) > 0 THEN to_timestamp(last_login_at) ELSE NULL END,
  0
FROM dedup, qpu
ON CONFLICT (id) DO UPDATE SET
  email = EXCLUDED.email,
  password_hash = EXCLUDED.password_hash,
  role = EXCLUDED.role,
  balance = EXCLUDED.balance,
  status = EXCLUDED.status,
  username = EXCLUDED.username,
  notes = EXCLUDED.notes,
  updated_at = NOW(),
  deleted_at = EXCLUDED.deleted_at,
  last_login_at = EXCLUDED.last_login_at`

const migrateAuthIdentitiesSQL = `
WITH identity_rows AS (
  SELECT id::bigint AS user_id, 'github' AS provider_type, 'default' AS provider_key, NULLIF(github_id, '') AS provider_subject, email, username FROM legacy_newapi.users
  UNION ALL
  SELECT id::bigint, 'linuxdo', 'default', NULLIF(linux_do_id, ''), email, username FROM legacy_newapi.users
  UNION ALL
  SELECT id::bigint, 'oidc', 'default', NULLIF(oidc_id, ''), email, username FROM legacy_newapi.users
  UNION ALL
  SELECT id::bigint, 'wechat', 'default', NULLIF(wechat_id, ''), email, username FROM legacy_newapi.users
),
dedup AS (
  SELECT *,
         ROW_NUMBER() OVER (
           PARTITION BY provider_type, provider_key, provider_subject
           ORDER BY user_id
         ) AS rn
  FROM identity_rows
  WHERE provider_subject IS NOT NULL
)
INSERT INTO auth_identities (user_id, provider_type, provider_key, provider_subject, verified_at, metadata, created_at, updated_at)
SELECT
  user_id,
  provider_type,
  provider_key,
  provider_subject,
  NOW(),
  jsonb_strip_nulls(jsonb_build_object('legacy_email', NULLIF(email, ''), 'legacy_username', NULLIF(username, ''))),
  NOW(),
  NOW()
FROM dedup
WHERE rn = 1
ON CONFLICT (provider_type, provider_key, provider_subject)
DO UPDATE SET user_id = EXCLUDED.user_id, metadata = EXCLUDED.metadata, updated_at = NOW()`

const migrateAPIKeysSQL = `
WITH ` + quotaPerUnitCTE + `,
prepared AS (
  SELECT
    t.*,
    NULLIF(btrim(split_part(COALESCE(NULLIF(t."group", ''), 'default'), ',', 1)), '') AS group_name
  FROM legacy_newapi.tokens t
)
INSERT INTO api_keys (
  id, user_id, "key", name, group_id, status, last_used_at, ip_whitelist,
  quota, quota_used, expires_at, created_at, updated_at, deleted_at
)
SELECT
  p.id::bigint,
  p.user_id::bigint,
  p.key,
  LEFT(COALESCE(NULLIF(p.name, ''), 'legacy-key-' || p.id::text), 100),
  g.id,
  CASE
    WHEN p.status = 1 THEN 'active'
    WHEN p.status = 3 THEN 'expired'
    ELSE 'disabled'
  END,
  CASE WHEN COALESCE(p.accessed_time, 0) > 0 THEN to_timestamp(p.accessed_time) ELSE NULL END,
  COALESCE((
    SELECT jsonb_agg(ip)
    FROM (
      SELECT NULLIF(btrim(x), '') AS ip
      FROM regexp_split_to_table(replace(COALESCE(p.allow_ips, ''), ',', E'\n'), E'[\s\n\r]+') AS x
    ) ips
    WHERE ip IS NOT NULL
  ), '[]'::jsonb),
  CASE WHEN p.unlimited_quota THEN 0 ELSE GREATEST(COALESCE(p.remain_quota, 0) + COALESCE(p.used_quota, 0), 0)::double precision / qpu.value END,
  GREATEST(COALESCE(p.used_quota, 0), 0)::double precision / qpu.value,
  CASE WHEN COALESCE(p.expired_time, -1) > 0 THEN to_timestamp(p.expired_time) ELSE NULL END,
  CASE WHEN COALESCE(p.created_time, 0) > 0 THEN to_timestamp(p.created_time) ELSE NOW() END,
  NOW(),
  p.deleted_at
FROM prepared p
CROSS JOIN qpu
LEFT JOIN groups g ON g.name = COALESCE(p.group_name, 'default') AND g.deleted_at IS NULL
JOIN users u ON u.id = p.user_id
WHERE COALESCE(p.key, '') <> ''
ON CONFLICT ("key") DO UPDATE SET
  user_id = EXCLUDED.user_id,
  name = EXCLUDED.name,
  group_id = EXCLUDED.group_id,
  status = EXCLUDED.status,
  last_used_at = EXCLUDED.last_used_at,
  ip_whitelist = EXCLUDED.ip_whitelist,
  quota = EXCLUDED.quota,
  quota_used = EXCLUDED.quota_used,
  expires_at = EXCLUDED.expires_at,
  updated_at = NOW(),
  deleted_at = EXCLUDED.deleted_at`

const migrateAccountsSQL = `
INSERT INTO accounts (
  id, name, notes, platform, type, credentials, extra, concurrency, priority,
  rate_multiplier, status, schedulable, created_at, updated_at, deleted_at
)
SELECT
  c.id::bigint,
  LEFT(COALESCE(NULLIF(c.name, ''), 'legacy-channel-' || c.id::text), 100),
  NULLIF(c.remark, ''),
  CASE
    WHEN c.type IN (14) THEN 'anthropic'
    WHEN c.type IN (24, 41) THEN 'gemini'
    ELSE 'openai'
  END,
  CASE WHEN c.type = 33 THEN 'bedrock' ELSE 'apikey' END,
  jsonb_strip_nulls(jsonb_build_object(
    'api_key', NULLIF(c.key, ''),
    'base_url', NULLIF(c.base_url, ''),
    'openai_organization', NULLIF(c.openai_organization, ''),
    'model_mapping', public.fuxi_migrate_safe_jsonb(c.model_mapping, '{}'::jsonb),
    'header_override', public.fuxi_migrate_safe_jsonb(c.header_override, '{}'::jsonb),
    'param_override', public.fuxi_migrate_safe_jsonb(c.param_override, '{}'::jsonb),
    'settings', public.fuxi_migrate_safe_jsonb(c.settings, '{}'::jsonb)
  )),
  jsonb_strip_nulls(jsonb_build_object(
    'legacy_newapi', jsonb_strip_nulls(jsonb_build_object(
      'id', c.id,
      'type', c.type,
      'models', NULLIF(c.models, ''),
      'group', NULLIF(c."group", ''),
      'test_model', c.test_model,
      'status_code_mapping', c.status_code_mapping,
      'other', NULLIF(c.other, ''),
      'other_info', public.fuxi_migrate_safe_jsonb(c.other_info, '{}'::jsonb),
      'setting', public.fuxi_migrate_safe_jsonb(c.setting, '{}'::jsonb),
      'channel_info', to_jsonb(c.channel_info),
      'tag', c.tag,
      'balance', c.balance,
      'used_quota', c.used_quota,
      'auto_ban', c.auto_ban,
      'response_time', c.response_time,
      'test_time', c.test_time
    ))
  )),
  3,
  COALESCE(c.priority::int, 50),
  1.0,
  CASE WHEN c.status = 1 THEN 'active' ELSE 'disabled' END,
  c.status = 1,
  CASE WHEN COALESCE(c.created_time, 0) > 0 THEN to_timestamp(c.created_time) ELSE NOW() END,
  NOW(),
  NULL
FROM legacy_newapi.channels c
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  notes = EXCLUDED.notes,
  platform = EXCLUDED.platform,
  type = EXCLUDED.type,
  credentials = EXCLUDED.credentials,
  extra = EXCLUDED.extra,
  priority = EXCLUDED.priority,
  status = EXCLUDED.status,
  schedulable = EXCLUDED.schedulable,
  updated_at = NOW()`

const migrateAccountGroupsSQL = `
WITH channel_groups AS (
  SELECT
    c.id::bigint AS account_id,
    NULLIF(btrim(group_name), '') AS group_name,
    COALESCE(c.priority::int, 50) AS priority
  FROM legacy_newapi.channels c
  CROSS JOIN LATERAL unnest(string_to_array(COALESCE(NULLIF(c."group", ''), 'default'), ',')) AS group_name
)
INSERT INTO account_groups (account_id, group_id, priority, created_at)
SELECT cg.account_id, g.id, MIN(cg.priority), NOW()
FROM channel_groups cg
JOIN accounts a ON a.id = cg.account_id
JOIN groups g ON g.name = COALESCE(cg.group_name, 'default') AND g.deleted_at IS NULL
GROUP BY cg.account_id, g.id
ON CONFLICT (account_id, group_id) DO UPDATE SET priority = EXCLUDED.priority`

const migrateRedeemCodesSQL = `
WITH ` + quotaPerUnitCTE + `
INSERT INTO redeem_codes (id, code, type, value, status, used_by, used_at, notes, created_at, group_id, validity_days)
SELECT
  r.id::bigint,
  r.key,
  'balance',
  COALESCE(r.quota, 0)::double precision / qpu.value,
  CASE
    WHEN r.status = 1 AND COALESCE(r.expired_time, 0) > 0 AND r.expired_time < EXTRACT(EPOCH FROM NOW()) THEN 'expired'
    WHEN r.status = 1 THEN 'unused'
    WHEN r.status = 3 THEN 'used'
    ELSE 'disabled'
  END,
  NULLIF(r.used_user_id, 0)::bigint,
  CASE WHEN COALESCE(r.redeemed_time, 0) > 0 THEN to_timestamp(r.redeemed_time) ELSE NULL END,
  concat_ws(E'\n', NULLIF(r.name, ''), CASE WHEN COALESCE(r.expired_time, 0) > 0 THEN 'legacy_expired_time=' || r.expired_time::text ELSE NULL END),
  CASE WHEN COALESCE(r.created_time, 0) > 0 THEN to_timestamp(r.created_time) ELSE NOW() END,
  NULL,
  30
FROM legacy_newapi.redemptions r, qpu
WHERE COALESCE(r.key, '') <> ''
ON CONFLICT (code) DO UPDATE SET
  type = EXCLUDED.type,
  value = EXCLUDED.value,
  status = EXCLUDED.status,
  used_by = EXCLUDED.used_by,
  used_at = EXCLUDED.used_at,
  notes = EXCLUDED.notes`

const migrateUsageLogsSQL = `
WITH ` + quotaPerUnitCTE + `,
prepared AS (
  SELECT
    l.*,
    NULLIF(btrim(split_part(COALESCE(NULLIF(l."group", ''), 'default'), ',', 1)), '') AS group_name
  FROM legacy_newapi.logs l
  WHERE l.type = 2
)
INSERT INTO usage_logs (
  user_id, api_key_id, account_id, request_id, model, requested_model, channel_id,
  group_id, input_tokens, output_tokens, total_cost, actual_cost, billing_type, stream,
  duration_ms, user_agent, ip_address, created_at
)
SELECT
  p.user_id::bigint,
  p.token_id::bigint,
  p.channel_id::bigint,
  LEFT(COALESCE(NULLIF(p.request_id, ''), 'legacy-log-' || p.id::text), 64),
  LEFT(COALESCE(NULLIF(p.model_name, ''), 'unknown'), 100),
  LEFT(NULLIF(p.model_name, ''), 100),
  p.channel_id::bigint,
  g.id,
  COALESCE(p.prompt_tokens, 0),
  COALESCE(p.completion_tokens, 0),
  ABS(COALESCE(p.quota, 0))::double precision / qpu.value,
  ABS(COALESCE(p.quota, 0))::double precision / qpu.value,
  0,
  p.is_stream,
  NULLIF(p.use_time, 0),
  NULL,
  NULLIF(p.ip, ''),
  CASE WHEN COALESCE(p.created_at, 0) > 0 THEN to_timestamp(p.created_at) ELSE NOW() END
FROM prepared p
CROSS JOIN qpu
JOIN users u ON u.id = p.user_id
JOIN api_keys ak ON ak.id = p.token_id
JOIN accounts a ON a.id = p.channel_id
LEFT JOIN groups g ON g.name = COALESCE(p.group_name, 'default') AND g.deleted_at IS NULL
WHERE p.token_id <> 0 AND p.channel_id <> 0
ON CONFLICT DO NOTHING`

const migratePaymentOrdersSQL = `
WITH prepared AS (
  SELECT
    t.*,
    COALESCE(NULLIF(t.payment_provider, ''), NULLIF(t.payment_method, ''), 'legacy') AS provider_key
  FROM legacy_newapi.top_ups t
)
INSERT INTO payment_orders (
  id, user_id, user_email, user_name, user_notes, amount, pay_amount, fee_rate,
  recharge_code, out_trade_no, payment_type, payment_trade_no, order_type,
  status, expires_at, paid_at, completed_at, failed_at, failed_reason,
  client_ip, src_host, src_url, created_at, updated_at
)
SELECT
  p.id::bigint,
  p.user_id::bigint,
  u.email,
  u.username,
  u.notes,
  CASE WHEN COALESCE(p.money, 0) > 0 THEN p.money ELSE COALESCE(p.amount, 0)::double precision END,
  CASE WHEN COALESCE(p.money, 0) > 0 THEN p.money ELSE COALESCE(p.amount, 0)::double precision END,
  0,
  LEFT(COALESCE(NULLIF(p.trade_no, ''), 'legacy-topup-' || p.id::text), 64),
  LEFT(COALESCE(NULLIF(p.trade_no, ''), 'legacy-topup-' || p.id::text), 64),
  LEFT(p.provider_key, 30),
  LEFT(COALESCE(NULLIF(p.trade_no, ''), 'legacy-topup-' || p.id::text), 128),
  'balance',
  CASE
    WHEN p.status = 'success' THEN 'COMPLETED'
    WHEN p.status = 'pending' THEN 'PENDING'
    WHEN p.status = 'expired' THEN 'EXPIRED'
    ELSE 'FAILED'
  END,
  CASE WHEN COALESCE(p.create_time, 0) > 0 THEN to_timestamp(p.create_time + 1800) ELSE NOW() + INTERVAL '30 minutes' END,
  CASE WHEN p.status = 'success' AND COALESCE(p.complete_time, 0) > 0 THEN to_timestamp(p.complete_time) ELSE NULL END,
  CASE WHEN p.status = 'success' AND COALESCE(p.complete_time, 0) > 0 THEN to_timestamp(p.complete_time) ELSE NULL END,
  CASE WHEN p.status NOT IN ('success', 'pending', 'expired') THEN NOW() ELSE NULL END,
  CASE WHEN p.status NOT IN ('success', 'pending', 'expired') THEN 'legacy status: ' || COALESCE(p.status, '') ELSE NULL END,
  '',
  'legacy.new-api',
  NULL,
  CASE WHEN COALESCE(p.create_time, 0) > 0 THEN to_timestamp(p.create_time) ELSE NOW() END,
  NOW()
FROM prepared p
JOIN users u ON u.id = p.user_id
ON CONFLICT (out_trade_no) WHERE out_trade_no <> ''
DO UPDATE SET
  status = EXCLUDED.status,
  paid_at = EXCLUDED.paid_at,
  completed_at = EXCLUDED.completed_at,
  failed_at = EXCLUDED.failed_at,
  failed_reason = EXCLUDED.failed_reason,
  updated_at = NOW()`
