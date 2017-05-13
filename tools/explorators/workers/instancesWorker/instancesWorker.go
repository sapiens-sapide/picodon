package instancesWorker

import (
	"context"
	expl "github.com/sapiens-sapide/picodon/tools/explorators"
)

type InstanceWorker struct {
	Backend     expl.Backend
	Context     context.Context
	Instance    expl.Instance
	IsConnected bool // whether worker is currently authenticated and connected to instance
}
