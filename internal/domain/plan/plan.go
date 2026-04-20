package plan

import "os"

type OperationType string

const (
	OpNote           OperationType = "note"
	OpEnsureEmptyDir OperationType = "ensure_empty_dir"
	OpEnsureExists   OperationType = "ensure_exists"
	OpMkdir          OperationType = "mkdir"
	OpWriteFile      OperationType = "write_file"
	OpRun            OperationType = "run"
	OpMigrateUp      OperationType = "migrate_up"
	OpMigrateDown    OperationType = "migrate_down"
	OpMigrateForce   OperationType = "migrate_force"
)

const (
	MigrateParamSourceURL   = "source_url"
	MigrateParamDatabaseURL = "database_url"
	MigrateParamSteps       = "steps"
	MigrateParamVersion     = "version"
)

type Operation struct {
	Type    OperationType
	Path    string
	Data    []byte
	Perm    os.FileMode
	Cmd     []string
	Message string
	Params  map[string]string
}

type Plan struct {
	CommandID   string
	Description string
	Ops         []Operation
	Warnings    []string
}
