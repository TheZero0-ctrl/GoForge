package newapp

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

const (
	defaultTemplateName = "default"
	defaultPackVersion  = "v1"
)

//go:embed templates/default/v1/manifest.json templates/default/v1/*.tmpl templates/default/v1/**/* templates/default/v1/.gitignore.tmpl templates/default/v1/migrations/.keep.tmpl
var templateFS embed.FS

type renderedFile struct {
	Path string
	Data []byte
}

type manifest struct {
	TemplateName string   `json:"template"`
	Version      string   `json:"version"`
	RequiredVars []string `json:"required_vars"`
	Files        []string `json:"files"`
}

type templateData struct {
	AppName          string
	ModulePath       string
	NormalizeAppName string
}

// TODO: maybe better way to do this?
func renderFiles(cfg Config) ([]renderedFile, error) {
	manifestPath := filepath.ToSlash(filepath.Join(
		"templates",
		defaultTemplateName,
		defaultPackVersion,
		"manifest.json",
	))

	rawManifest, err := templateFS.ReadFile(manifestPath)

	if err != nil {
		return nil, fmt.Errorf("read manifest %q: %w", manifestPath, err)
	}

	var m manifest

	if err := json.Unmarshal(rawManifest, &m); err != nil {
		return nil, fmt.Errorf("parse manifest %q: %w", manifestPath, err)
	}

	files := append([]string(nil), m.Files...)
	sort.Strings(files)

	data := templateData{
		AppName:          cfg.AppName,
		NormalizeAppName: cfg.NormalizeAppName,
		ModulePath:       cfg.ModulePath,
	}

	out := make([]renderedFile, 0, len(files))

	for _, relPath := range files {
		tplPath := filepath.ToSlash(filepath.Join(
			"templates",
			defaultTemplateName,
			defaultPackVersion,
			relPath,
		))
		tplBytes, err := templateFS.ReadFile(tplPath)
		if err != nil {
			return nil, fmt.Errorf("read template %q: %w", relPath, err)
		}
		tpl, err := template.New(filepath.Base(relPath)).
			Option("missingkey=error").
			Parse(string(tplBytes))

		if err != nil {
			return nil, fmt.Errorf("parse template %q: %w", relPath, err)
		}

		var buf bytes.Buffer
		if err := tpl.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("render template %q: %w", relPath, err)
		}
		outPath := strings.TrimSuffix(relPath, ".tmpl")

		out = append(out, renderedFile{
			Path: outPath,
			Data: buf.Bytes(),
		})
	}

	return out, nil
}
