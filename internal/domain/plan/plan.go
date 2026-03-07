package plan

import "os"

type OperationType string

const (
	OpNote           OperationType = "note"
	OpEnsureEmptyDir OperationType = "ensure_empty_dir"
	OpMkdir          OperationType = "mkdir"
	OpWriteFile      OperationType = "write_file"
	OpRun            OperationType = "run"
)

type Operation struct {
	Type    OperationType
	Path    string
	Data    []byte
	Perm    os.FileMode
	Cmd     []string
	Message string
}

type Plan struct {
	CommandID   string
	Description string
	Ops         []Operation
	Warnings    []string
}
