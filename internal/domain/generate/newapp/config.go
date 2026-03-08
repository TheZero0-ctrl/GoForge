package newapp

import (
	"fmt"
	"regexp"
	"strings"
)

var nonURLChars = regexp.MustCompile(`(?i)[^a-z0-9\-_]+`)

const defaultPort = 3000

type Config struct {
	AppName          string
	ModulePath       string
	NormalizeAppName string
	SkipGit          bool
	SkipTidy         bool
}

func ParseConfig(args []string, p Params) (Config, error) {
	if len(args) != 1 {
		return Config{}, fmt.Errorf("new requires exactly one argument: <app-name>")
	}

	appName := strings.TrimSpace(args[0])
	if appName == "" {
		return Config{}, fmt.Errorf("app name cannot be empty")
	}

	normalizeAppName := sanitize(appName, "_")

	// TODO: maybe latter we nedd to validate appName and module?

	module := p.Param("module")
	if module == "" {
		module = appName
	}

	return Config{
		AppName:          appName,
		ModulePath:       module,
		NormalizeAppName: normalizeAppName,
		SkipGit:          p.BoolParam("skip-git"),
		SkipTidy:         p.BoolParam("skip-tidy"),
	}, nil
}

func sanitize(s, sep string) string {
	return nonURLChars.ReplaceAllString(s, sep)
}
