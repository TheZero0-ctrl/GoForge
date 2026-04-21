package resource

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"text/template"
)

const (
	dataTemplatePath     = "templates/default/v1/data.go.tmpl"
	handlersTemplatePath = "templates/default/v1/handlers.go.tmpl"
)

//go:embed templates/default/v1/*.tmpl
var templateFS embed.FS

func renderDataFile(ctx templateContext) ([]byte, error) {
	tpl, err := loadTemplate(dataTemplatePath, "data", template.FuncMap{"add": func(a, b int) int { return a + b }})
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := tpl.Execute(&out, ctx); err != nil {
		return nil, fmt.Errorf("render data template: %w", err)
	}

	src, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated data file: %w", err)
	}

	return src, nil
}

func renderHandlersFile(ctx templateContext) ([]byte, error) {
	tpl, err := loadTemplate(handlersTemplatePath, "handlers", nil)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := tpl.Execute(&out, ctx); err != nil {
		return nil, fmt.Errorf("render handler template: %w", err)
	}

	src, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated handler file: %w", err)
	}

	return src, nil
}

func loadTemplate(path, name string, funcs template.FuncMap) (*template.Template, error) {
	raw, err := templateFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template %q: %w", path, err)
	}

	tpl := template.New(name).Option("missingkey=error")
	if funcs != nil {
		tpl = tpl.Funcs(funcs)
	}

	tpl, err = tpl.Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse template %q: %w", path, err)
	}

	return tpl, nil
}
