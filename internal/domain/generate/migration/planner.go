package migration

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"goforge/internal/domain/params"
	"goforge/internal/domain/plan"
)

var migrationNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
var createPattern = regexp.MustCompile(`^create_([a-z][a-z0-9_]*)$`)
var addPattern = regexp.MustCompile(`^add_([a-z][a-z0-9_]*)_to_([a-z][a-z0-9_]*)$`)
var removePattern = regexp.MustCompile(`^remove_([a-z][a-z0-9_]*)_from_([a-z][a-z0-9_]*)$`)

const timestampFormat = "20060102150405"

type patternKind string

const (
	patternCustom patternKind = "custom"
	patternCreate patternKind = "create"
	patternAdd    patternKind = "add"
	patternRemove patternKind = "remove"
)

type parsedPattern struct {
	kind   patternKind
	table  string
	column string
}

type fieldSpec struct {
	name    string
	sqlType string
}

func Validate(args []string, _ params.Params) error {
	if len(args) < 1 {
		return fmt.Errorf("generate migration requires at least one argument: <name>")
	}

	name := strings.TrimSpace(args[0])
	if name == "" {
		return fmt.Errorf("migration name cannot be empty")
	}

	if !migrationNamePattern.MatchString(name) {
		return fmt.Errorf("migration name %q is invalid; use snake_case", name)
	}

	pattern := parsePattern(name)
	if pattern.kind == patternCustom {
		return nil
	}

	_, err := parseFields(args[1:])
	if err != nil {
		return err
	}

	return nil
}

func Plan(_ context.Context, args []string, _ params.Params) (plan.Plan, error) {
	name := strings.TrimSpace(args[0])
	pattern := parsePattern(name)
	fields, err := parseFields(args[1:])
	if err != nil {
		return plan.Plan{}, err
	}

	upData, downData := renderMigration(pattern, fields)

	version := time.Now().UTC().Format(timestampFormat)
	baseName := fmt.Sprintf("%s_%s", version, name)

	upPath := filepath.Join("migrations", baseName+".up.sql")
	downPath := filepath.Join("migrations", baseName+".down.sql")

	return plan.Plan{
		CommandID:   "generate:migration",
		Description: "Generate migration files",
		Ops: []plan.Operation{
			{Type: plan.OpMkdir, Path: "migrations"},
			{Type: plan.OpWriteFile, Path: upPath, Data: upData},
			{Type: plan.OpWriteFile, Path: downPath, Data: downData},
		},
	}, nil
}

func parsePattern(name string) parsedPattern {
	if m := createPattern.FindStringSubmatch(name); len(m) == 2 {
		return parsedPattern{kind: patternCreate, table: m[1]}
	}

	if m := addPattern.FindStringSubmatch(name); len(m) == 3 {
		return parsedPattern{kind: patternAdd, column: m[1], table: m[2]}
	}

	if m := removePattern.FindStringSubmatch(name); len(m) == 3 {
		return parsedPattern{kind: patternRemove, column: m[1], table: m[2]}
	}

	return parsedPattern{kind: patternCustom}
}

func parseFields(tokens []string) (map[string]fieldSpec, error) {
	fields := make(map[string]fieldSpec, len(tokens))
	for _, token := range tokens {
		parts := strings.SplitN(strings.TrimSpace(token), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid field %q; expected <name>:<type>", token)
		}

		name := strings.TrimSpace(parts[0])
		if !migrationNamePattern.MatchString(name) {
			return nil, fmt.Errorf("invalid field name %q; use snake_case", name)
		}

		sqlType := toSQLType(strings.TrimSpace(parts[1]))
		if sqlType == "" {
			return nil, fmt.Errorf("unsupported field type %q (supported: string, int, int64, bool, float64, time, string[], runtime)", strings.TrimSpace(parts[1]))
		}

		fields[name] = fieldSpec{name: name, sqlType: sqlType}
	}

	return fields, nil
}

func renderMigration(pattern parsedPattern, fields map[string]fieldSpec) ([]byte, []byte) {
	switch pattern.kind {
	case patternCreate:
		return renderCreateTable(pattern.table, fields)
	case patternAdd:
		return renderAddColumn(pattern.table, pattern.column, fields)
	case patternRemove:
		return renderRemoveColumn(pattern.table, pattern.column, fields)
	default:
		return []byte{}, []byte{}
	}
}

func renderCreateTable(table string, fields map[string]fieldSpec) ([]byte, []byte) {
	columns := []string{"id bigserial PRIMARY KEY"}

	names := sortedFieldNames(fields)
	for _, name := range names {
		spec := fields[name]
		columns = append(columns, fmt.Sprintf("%s %s", spec.name, spec.sqlType))
	}

	up := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n);\n", table, strings.Join(columns, ",\n  "))
	down := fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", table)

	return []byte(up), []byte(down)
}

func renderAddColumn(table, column string, fields map[string]fieldSpec) ([]byte, []byte) {
	spec, ok := fields[column]
	if !ok {
		return []byte{}, []byte{}
	}

	up := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s;\n", table, column, spec.sqlType)
	down := fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s;\n", table, column)

	return []byte(up), []byte(down)
}

func renderRemoveColumn(table, column string, fields map[string]fieldSpec) ([]byte, []byte) {
	up := fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s;\n", table, column)

	spec, ok := fields[column]
	if !ok {
		return []byte(up), []byte{}
	}

	down := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s;\n", table, column, spec.sqlType)
	return []byte(up), []byte(down)
}

func sortedFieldNames(fields map[string]fieldSpec) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func toSQLType(fieldType string) string {
	switch fieldType {
	case "string":
		return "text"
	case "int":
		return "integer"
	case "int64":
		return "bigint"
	case "bool":
		return "boolean"
	case "float64":
		return "double precision"
	case "time":
		return "timestamp with time zone"
	case "string[]":
		return "text[]"
	case "runtime":
		return "integer"
	default:
		return ""
	}
}
