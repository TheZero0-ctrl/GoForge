package fielddsl

import (
	"fmt"
	"strings"

	"goforge/internal/domain/naming"
)

type Field struct {
	Name string
	Type TypeSpec
}

func (f Field) GoName() string {
	return naming.ToPascal(f.Name)
}

func (f Field) JSONName() string {
	return f.Name
}

func (f Field) DBName() string {
	return f.Name
}

func ParseToken(token string) (Field, error) {
	parts := strings.SplitN(strings.TrimSpace(token), ":", 2)
	if len(parts) != 2 {
		return Field{}, fmt.Errorf("invalid field %q; expected <name>:<type>", token)
	}

	name := naming.ToSnake(strings.TrimSpace(parts[0]))
	if name == "" {
		return Field{}, fmt.Errorf("field name cannot be empty")
	}

	if !naming.IsSnakeIdentifier(name) {
		return Field{}, fmt.Errorf("invalid field name %q; use letters, numbers, and underscores only", strings.TrimSpace(parts[0]))
	}

	typeKey := strings.TrimSpace(parts[1])
	if typeKey == "" {
		return Field{}, fmt.Errorf("field type cannot be empty")
	}

	typeSpec, ok := LookupType(typeKey)
	if !ok {
		return Field{}, fmt.Errorf("unsupported field type %q (supported: %s)", typeKey, SupportedTypesHint())
	}

	return Field{Name: name, Type: typeSpec}, nil
}

func ParseMany(tokens []string) ([]Field, error) {
	fields := make([]Field, 0, len(tokens))
	seen := make(map[string]struct{}, len(tokens))

	for _, token := range tokens {
		field, err := ParseToken(token)
		if err != nil {
			return nil, err
		}

		if _, exists := seen[field.Name]; exists {
			return nil, fmt.Errorf("duplicate field name %q", field.Name)
		}

		seen[field.Name] = struct{}{}
		fields = append(fields, field)
	}

	return fields, nil
}
