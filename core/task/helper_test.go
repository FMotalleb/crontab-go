package task

import (
	"context"
	"testing"

	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

// TestPopulateVarsClonesSharedMap guards against mutating the vars map shared
// through the context by concurrent task executions (data race).
func TestPopulateVarsClonesSharedMap(t *testing.T) {
	shared := map[string]string{"shared": "value"}
	parent := context.WithValue(context.Background(), ctxutils.Vars, shared)

	child := populateVars(parent, &config.Task{
		Vars: map[string]string{"local": "v1"},
	})

	childVars, ok := child.Value(ctxutils.Vars).(map[string]string)
	if !ok {
		t.Fatal("populateVars did not store a vars map in the child context")
	}
	if childVars["local"] != "v1" {
		t.Fatalf("expected the local var to be set in the child context, got %v", childVars)
	}
	if _, mutated := shared["local"]; mutated {
		t.Fatal("populateVars mutated the parent context's shared vars map")
	}
}
