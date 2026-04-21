package fielddsl

import (
	"sort"
	"strings"
)

type TypeSpec struct {
	Key         string
	GoType      string
	SQLType     string
	NeedsHelper bool
	Description string
}

var supportedTypes = []TypeSpec{
	{Key: "string", GoType: "string", SQLType: "TEXT", Description: "UTF-8 text"},
	{Key: "int", GoType: "int", SQLType: "INTEGER", Description: "Machine-sized integer"},
	{Key: "int64", GoType: "int64", SQLType: "BIGINT", Description: "64-bit integer"},
	{Key: "bool", GoType: "bool", SQLType: "BOOLEAN", Description: "True/false value"},
	{Key: "float64", GoType: "float64", SQLType: "DOUBLE PRECISION", Description: "64-bit floating point"},
	{Key: "time", GoType: "time.Time", SQLType: "TIMESTAMP WITH TIME ZONE", Description: "Timestamp with timezone"},
	{Key: "string[]", GoType: "[]string", SQLType: "TEXT[]", Description: "Array of text values"},
}

var supportedByKey map[string]TypeSpec

func init() {
	supportedByKey = make(map[string]TypeSpec, len(supportedTypes))
	for _, spec := range supportedTypes {
		supportedByKey[spec.Key] = spec
	}
}

func LookupType(raw string) (TypeSpec, bool) {
	key := strings.ToLower(strings.TrimSpace(raw))
	spec, ok := supportedByKey[key]
	return spec, ok
}

func SupportedTypes() []TypeSpec {
	out := make([]TypeSpec, len(supportedTypes))
	copy(out, supportedTypes)
	return out
}

func SupportedTypeKeys() []string {
	keys := make([]string, 0, len(supportedTypes))
	for _, spec := range supportedTypes {
		keys = append(keys, spec.Key)
	}
	sort.Strings(keys)
	return keys
}

func SupportedTypesHint() string {
	return strings.Join(SupportedTypeKeys(), ", ")
}
