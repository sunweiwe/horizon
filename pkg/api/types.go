package api

const (
	StatusOK = "ok"
)

const (
	WorkspaceNone = ""
	ClusterNone   = ""
)

type ListResult struct {
	Items      []interface{} `json:"items"`
	TotalItems int           `json:"totalItems"`
}
