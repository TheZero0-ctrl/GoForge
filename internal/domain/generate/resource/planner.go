package resource

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"goforge/internal/domain/generate/fielddsl"
	"goforge/internal/domain/generate/migration"
	"goforge/internal/domain/naming"
	"goforge/internal/domain/params"
	"goforge/internal/domain/plan"
)

type templateField struct {
	NameSnake        string
	NamePascal       string
	NameCamel        string
	DataGoType       string
	HandlerGoType    string
	HandlerPtrType   string
	InsertArgExpr    string
	ScanArgExpr      string
	UpdateAssignExpr string
	TypeKey          string
	IsStringArray    bool
	IsBool           bool
	IsString         bool
	IsTime           bool
	IsNumeric        bool
}

type templateContext struct {
	ModulePath       string
	Singular         string
	Plural           string
	SingularPascal   string
	PluralPascal     string
	SingularCamel    string
	PluralLowerCamel string
	Fields           []templateField
	DBColumnsCSV     string
	InsertValuesCSV  string
	UpdateSetSQL     string
	GetScanArgsCSV   string
	InsertArgsCSV    string
	UpdateArgsCSV    string
	HasStringArray   bool
	HasTimeField     bool
}

func Validate(args []string, _ params.Params) error {
	_, _, err := parseInput(args)
	if err != nil {
		return err
	}

	required := []string{
		"go.mod",
		"cmd/api/main.go",
		"cmd/api/routes.go",
		"internal/data/models.go",
		"internal/validator/validator.go",
		"migrations",
	}

	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("missing %s; run from a GoForge app root", path)
			}
			return fmt.Errorf("check %s: %w", path, err)
		}
	}

	return nil
}

func Plan(ctx context.Context, args []string, _ params.Params) (plan.Plan, error) {
	names, fields, err := parseInput(args)
	if err != nil {
		return plan.Plan{}, err
	}

	if conflictPath, hasConflict, err := findCreateMigrationConflict("migrations", names.Plural); err != nil {
		return plan.Plan{}, err
	} else if hasConflict {
		return plan.Plan{
			CommandID:   "generate:resource",
			Description: "Generate resource files and migration",
			Ops: []plan.Operation{{
				Type:    plan.OpEnsureNotExists,
				Path:    conflictPath,
				Message: fmt.Sprintf("conflict: create migration for %q already exists (%s)", names.Plural, conflictPath),
			}},
		}, nil
	}

	modulePath, err := readModulePath("go.mod")
	if err != nil {
		return plan.Plan{}, err
	}

	tplCtx := buildTemplateContext(modulePath, names, fields)

	dataBytes, err := renderDataFile(tplCtx)
	if err != nil {
		return plan.Plan{}, err
	}

	handlerBytes, err := renderHandlersFile(tplCtx)
	if err != nil {
		return plan.Plan{}, err
	}

	ops := []plan.Operation{
		{Type: plan.OpEnsureExists, Path: "cmd/api/routes.go", Message: "missing cmd/api/routes.go; run from GoForge app root"},
		{Type: plan.OpEnsureExists, Path: "internal/data/models.go", Message: "missing internal/data/models.go; run from GoForge app root"},
		{Type: plan.OpEnsureExists, Path: "migrations", Message: "missing migrations; run from GoForge app root"},
		{Type: plan.OpWriteFile, Path: filepath.Join("internal", "data", names.Plural+".go"), Data: dataBytes},
		{Type: plan.OpWriteFile, Path: filepath.Join("cmd", "api", names.Plural+".go"), Data: handlerBytes},
	}

	gofmtTargets := []string{
		filepath.Join("internal", "data", names.Plural+".go"),
		filepath.Join("cmd", "api", names.Plural+".go"),
	}

	routesPath := filepath.Join("cmd", "api", "routes.go")
	routesCurrent, err := os.ReadFile(routesPath)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("read %s: %w", routesPath, err)
	}

	routesUpdated, routesChanged, err := patchRoutesFile(routesCurrent, names)
	if err != nil {
		return plan.Plan{}, err
	}
	if routesChanged {
		ops = append(ops, plan.Operation{Type: plan.OpUpdateFile, Path: routesPath, Data: routesUpdated})
		gofmtTargets = append(gofmtTargets, routesPath)
	}

	modelsPath := filepath.Join("internal", "data", "models.go")
	modelsCurrent, err := os.ReadFile(modelsPath)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("read %s: %w", modelsPath, err)
	}

	modelsUpdated, modelsChanged, err := patchModelsFile(modelsCurrent, names)
	if err != nil {
		return plan.Plan{}, err
	}
	if modelsChanged {
		ops = append(ops, plan.Operation{Type: plan.OpUpdateFile, Path: modelsPath, Data: modelsUpdated})
		gofmtTargets = append(gofmtTargets, modelsPath)
	}

	migrationArgs := make([]string, 0, len(fields)+1)
	migrationArgs = append(migrationArgs, "create_"+names.Plural)
	for _, field := range fields {
		migrationArgs = append(migrationArgs, fmt.Sprintf("%s:%s", field.Name, field.Type.Key))
	}

	migrationPlan, err := migration.Plan(ctx, migrationArgs, nil)
	if err != nil {
		return plan.Plan{}, err
	}
	ops = append(ops, migrationPlan.Ops...)

	sort.Strings(gofmtTargets)
	ops = append(ops, plan.Operation{Type: plan.OpRun, Cmd: append([]string{"gofmt", "-w"}, gofmtTargets...)})

	return plan.Plan{
		CommandID:   "generate:resource",
		Description: "Generate resource files and migration",
		Ops:         ops,
	}, nil
}

func findCreateMigrationConflict(migrationsDir, plural string) (string, bool, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("read %s: %w", migrationsDir, err)
	}

	upSuffix := "_create_" + plural + ".up.sql"
	downSuffix := "_create_" + plural + ".down.sql"
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, upSuffix) || strings.HasSuffix(name, downSuffix) {
			return filepath.Join(migrationsDir, name), true, nil
		}
	}

	return "", false, nil
}

func parseInput(args []string) (naming.ResourceNames, []fielddsl.Field, error) {
	if len(args) < 1 {
		return naming.ResourceNames{}, nil, fmt.Errorf("generate resource requires arguments: <name> <field:type>...")
	}

	names, err := naming.NormalizeResourceName(args[0])
	if err != nil {
		return naming.ResourceNames{}, nil, err
	}

	if len(args) < 2 {
		return naming.ResourceNames{}, nil, fmt.Errorf("generate resource requires at least one field: <field:type>")
	}

	fields, err := fielddsl.ParseMany(args[1:])
	if err != nil {
		return naming.ResourceNames{}, nil, err
	}

	return names, fields, nil
}

func readModulePath(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") {
			module := strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
			if module == "" {
				return "", fmt.Errorf("module path is empty in %s", path)
			}
			return module, nil
		}
	}

	return "", fmt.Errorf("module path not found in %s", path)
}

func buildTemplateContext(modulePath string, names naming.ResourceNames, fields []fielddsl.Field) templateContext {
	templateFields := make([]templateField, 0, len(fields))
	columns := make([]string, 0, len(fields))
	insertValues := make([]string, 0, len(fields))
	updateSet := make([]string, 0, len(fields))
	getScanArgs := []string{"&" + names.SingularCamel + ".ID", "&" + names.SingularCamel + ".CreatedAt", "&" + names.SingularCamel + ".Version"}
	insertArgs := make([]string, 0, len(fields))
	updateArgs := make([]string, 0, len(fields)+2)

	hasStringArray := false
	hasTimeField := false

	for idx, field := range fields {
		pascal := field.GoName()
		camel := naming.ToLowerCamel(pascal)

		dataType := field.Type.GoType
		handlerType := field.Type.GoType
		handlerPtr := "*" + handlerType

		insertExpr := names.SingularCamel + "." + pascal
		scanExpr := "&" + names.SingularCamel + "." + pascal
		updateAssign := names.SingularCamel + "." + pascal + " = *input." + pascal
		if field.Type.Key == "string[]" {
			hasStringArray = true
			insertExpr = "pq.Array(" + names.SingularCamel + "." + pascal + ")"
			scanExpr = "pq.Array(&" + names.SingularCamel + "." + pascal + ")"
			handlerPtr = handlerType
			updateAssign = names.SingularCamel + "." + pascal + " = input." + pascal
		}

		if field.Type.Key == "time" {
			hasTimeField = true
		}

		templateFields = append(templateFields, templateField{
			NameSnake:        field.Name,
			NamePascal:       pascal,
			NameCamel:        camel,
			DataGoType:       dataType,
			HandlerGoType:    handlerType,
			HandlerPtrType:   handlerPtr,
			InsertArgExpr:    insertExpr,
			ScanArgExpr:      scanExpr,
			UpdateAssignExpr: updateAssign,
			TypeKey:          field.Type.Key,
			IsStringArray:    field.Type.Key == "string[]",
			IsBool:           field.Type.Key == "bool",
			IsString:         field.Type.Key == "string",
			IsTime:           field.Type.Key == "time",
			IsNumeric:        field.Type.Key == "int" || field.Type.Key == "int64" || field.Type.Key == "float64",
		})

		columns = append(columns, field.Name)
		insertValues = append(insertValues, fmt.Sprintf("$%d", idx+1))
		updateSet = append(updateSet, fmt.Sprintf("%s = $%d", field.Name, idx+1))
		getScanArgs = append(getScanArgs, scanExpr)
		insertArgs = append(insertArgs, insertExpr)
		updateArgs = append(updateArgs, insertExpr)
	}

	updateArgs = append(updateArgs, names.SingularCamel+".ID", names.SingularCamel+".Version")

	return templateContext{
		ModulePath:       modulePath,
		Singular:         names.Singular,
		Plural:           names.Plural,
		SingularPascal:   names.SingularPascal,
		PluralPascal:     names.PluralPascal,
		SingularCamel:    names.SingularCamel,
		PluralLowerCamel: names.PluralCamel,
		Fields:           templateFields,
		DBColumnsCSV:     strings.Join(columns, ", "),
		InsertValuesCSV:  strings.Join(insertValues, ", "),
		UpdateSetSQL:     strings.Join(updateSet, ", "),
		GetScanArgsCSV:   strings.Join(getScanArgs, ",\n\t\t"),
		InsertArgsCSV:    strings.Join(insertArgs, ",\n\t\t"),
		UpdateArgsCSV:    strings.Join(updateArgs, ",\n\t\t"),
		HasStringArray:   hasStringArray,
		HasTimeField:     hasTimeField,
	}
}

func patchRoutesFile(src []byte, names naming.ResourceNames) ([]byte, bool, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "cmd/api/routes.go", src, 0)
	if err != nil {
		return nil, false, fmt.Errorf("parse cmd/api/routes.go: %w", err)
	}

	routesFunc := findFunc(file, "routes")
	if routesFunc == nil || routesFunc.Body == nil {
		return nil, false, fmt.Errorf("could not find routes() function in cmd/api/routes.go")
	}

	type routeSpec struct {
		methodConst string
		path        string
		handler     string
	}

	desired := []routeSpec{
		{methodConst: "MethodGet", path: "/v1/" + names.Plural, handler: "list" + names.PluralPascal + "Handler"},
		{methodConst: "MethodPost", path: "/v1/" + names.Plural, handler: "create" + names.SingularPascal + "Handler"},
		{methodConst: "MethodGet", path: "/v1/" + names.Plural + "/:id", handler: "show" + names.SingularPascal + "Handler"},
		{methodConst: "MethodPatch", path: "/v1/" + names.Plural + "/:id", handler: "update" + names.SingularPascal + "Handler"},
		{methodConst: "MethodDelete", path: "/v1/" + names.Plural + "/:id", handler: "delete" + names.SingularPascal + "Handler"},
	}

	existingByMethodPath := map[string]string{}
	for _, stmt := range routesFunc.Body.List {
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}
		call, ok := exprStmt.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 3 {
			continue
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "HandlerFunc" {
			continue
		}
		xIdent, ok := sel.X.(*ast.Ident)
		if !ok || xIdent.Name != "router" {
			continue
		}
		method, ok := call.Args[0].(*ast.SelectorExpr)
		if !ok {
			continue
		}
		httpPkg, ok := method.X.(*ast.Ident)
		if !ok || httpPkg.Name != "http" {
			continue
		}
		pathLit, ok := call.Args[1].(*ast.BasicLit)
		if !ok {
			continue
		}
		handlerSel, ok := call.Args[2].(*ast.SelectorExpr)
		if !ok {
			continue
		}
		appIdent, ok := handlerSel.X.(*ast.Ident)
		if !ok || appIdent.Name != "app" {
			continue
		}

		methodPathKey := method.Sel.Name + ":" + pathLit.Value
		handlerName := handlerSel.Sel.Name
		if priorHandler, exists := existingByMethodPath[methodPathKey]; exists && priorHandler != handlerName {
			return nil, false, fmt.Errorf("route conflict for %s: existing handler %q differs from %q", methodPathKey, priorHandler, handlerName)
		}
		existingByMethodPath[methodPathKey] = handlerName
	}

	toAdd := make([]ast.Stmt, 0, len(desired))
	for _, route := range desired {
		methodPathKey := route.methodConst + ":\"" + route.path + "\""
		if existingHandler, ok := existingByMethodPath[methodPathKey]; ok {
			if existingHandler != route.handler {
				return nil, false, fmt.Errorf("route conflict for %s: existing handler %q differs from %q", methodPathKey, existingHandler, route.handler)
			}
			continue
		}

		toAdd = append(toAdd, &ast.ExprStmt{X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{X: ast.NewIdent("router"), Sel: ast.NewIdent("HandlerFunc")},
			Args: []ast.Expr{
				&ast.SelectorExpr{X: ast.NewIdent("http"), Sel: ast.NewIdent(route.methodConst)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", route.path)},
				&ast.SelectorExpr{X: ast.NewIdent("app"), Sel: ast.NewIdent(route.handler)},
			},
		}})
	}

	if len(toAdd) == 0 {
		return src, false, nil
	}

	insertAt := len(routesFunc.Body.List)
	for i, stmt := range routesFunc.Body.List {
		if _, ok := stmt.(*ast.ReturnStmt); ok {
			insertAt = i
			break
		}
	}

	body := append([]ast.Stmt{}, routesFunc.Body.List[:insertAt]...)
	body = append(body, toAdd...)
	body = append(body, routesFunc.Body.List[insertAt:]...)
	routesFunc.Body.List = body

	var out bytes.Buffer
	if err := format.Node(&out, fset, file); err != nil {
		return nil, false, fmt.Errorf("format updated cmd/api/routes.go: %w", err)
	}

	return out.Bytes(), true, nil
}

func patchModelsFile(src []byte, names naming.ResourceNames) ([]byte, bool, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "internal/data/models.go", src, 0)
	if err != nil {
		return nil, false, fmt.Errorf("parse internal/data/models.go: %w", err)
	}

	modelsStruct := findTypeStruct(file, "Models")
	if modelsStruct == nil {
		return nil, false, fmt.Errorf("could not find type Models struct in internal/data/models.go")
	}

	fieldExists := false
	for _, field := range modelsStruct.Fields.List {
		for _, name := range field.Names {
			if name.Name == names.PluralPascal {
				typeIdent, ok := field.Type.(*ast.Ident)
				if !ok || typeIdent.Name != names.SingularPascal+"Model" {
					return nil, false, fmt.Errorf("models conflict: field %s exists with incompatible type", names.PluralPascal)
				}
				fieldExists = true
				break
			}
		}
	}

	changed := false
	if !fieldExists {
		modelsStruct.Fields.List = append(modelsStruct.Fields.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(names.PluralPascal)},
			Type:  ast.NewIdent(names.SingularPascal + "Model"),
		})
		changed = true
	}

	newModels := findFunc(file, "NewModels")
	if newModels == nil || newModels.Body == nil {
		return nil, false, fmt.Errorf("could not find NewModels function in internal/data/models.go")
	}

	returnLit := findReturnModelsComposite(newModels)
	if returnLit == nil {
		return nil, false, fmt.Errorf("could not find return Models literal in NewModels")
	}

	keyExists := false
	for _, elt := range returnLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if ok && key.Name == names.PluralPascal {
			if !isExpectedModelValue(kv.Value, names.SingularPascal) {
				return nil, false, fmt.Errorf("models conflict: NewModels already wires %s with incompatible value", names.PluralPascal)
			}
			keyExists = true
			break
		}
	}

	if !keyExists {
		returnLit.Elts = append(returnLit.Elts, &ast.KeyValueExpr{
			Key: ast.NewIdent(names.PluralPascal),
			Value: &ast.CompositeLit{
				Type: ast.NewIdent(names.SingularPascal + "Model"),
				Elts: []ast.Expr{&ast.KeyValueExpr{Key: ast.NewIdent("DB"), Value: ast.NewIdent("db")}},
			},
		})
		changed = true
	}

	if !changed {
		return src, false, nil
	}

	var out bytes.Buffer
	if err := format.Node(&out, fset, file); err != nil {
		return nil, false, fmt.Errorf("format updated internal/data/models.go: %w", err)
	}

	return out.Bytes(), true, nil
}

func findFunc(file *ast.File, name string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name != nil && fn.Name.Name == name {
			return fn
		}
	}
	return nil
}

func findTypeStruct(file *ast.File, name string) *ast.StructType {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name == nil || ts.Name.Name != name {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if ok {
				return st
			}
		}
	}
	return nil
}

func findReturnModelsComposite(fn *ast.FuncDecl) *ast.CompositeLit {
	for _, stmt := range fn.Body.List {
		ret, ok := stmt.(*ast.ReturnStmt)
		if !ok || len(ret.Results) == 0 {
			continue
		}
		lit, ok := ret.Results[0].(*ast.CompositeLit)
		if !ok {
			continue
		}
		ident, ok := lit.Type.(*ast.Ident)
		if ok && ident.Name == "Models" {
			return lit
		}
	}
	return nil
}

func isExpectedModelValue(expr ast.Expr, singularPascal string) bool {
	composite, ok := expr.(*ast.CompositeLit)
	if !ok {
		return false
	}

	typeIdent, ok := composite.Type.(*ast.Ident)
	if !ok || typeIdent.Name != singularPascal+"Model" {
		return false
	}

	if len(composite.Elts) != 1 {
		return false
	}

	kv, ok := composite.Elts[0].(*ast.KeyValueExpr)
	if !ok {
		return false
	}

	key, ok := kv.Key.(*ast.Ident)
	if !ok || key.Name != "DB" {
		return false
	}

	value, ok := kv.Value.(*ast.Ident)
	if !ok || value.Name != "db" {
		return false
	}

	return true
}
