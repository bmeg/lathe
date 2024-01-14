package workflow

import (
	"fmt"
	"log"

	"github.com/bmeg/flame"
)

/*****/

type WorkflowFileCheck struct {
	File DataFile
}

func (ws *WorkflowFileCheck) Process(key string, status []*WorkflowStatus) flame.KeyValue[string, *WorkflowStatus] {
	dryRun := false
	for _, i := range status {
		if i.Status != STATUS_OK {
			return flame.KeyValue[string, *WorkflowStatus]{Key: key, Value: i}
		}
		if i.DryRun {
			dryRun = true
		}
	}
	output := &WorkflowStatus{DryRun: dryRun}
	log.Printf("Checking for file: %s\n", ws.File.Abs())
	if !PathExists(ws.File.Abs()) {
		output.Status = STATUS_FAIL
		log.Printf("Missing file: %s\n", ws.File.Abs())
	} else {
		output.Status = STATUS_OK
	}
	return flame.KeyValue[string, *WorkflowStatus]{Key: key, Value: output}
}

func (ws *WorkflowFileCheck) IsGenerator() bool {
	return true
}

func (ws *WorkflowFileCheck) GetInputs() map[string]DataFile {
	out := map[string]DataFile{}
	return out
}

func (ws *WorkflowFileCheck) GetOutputs() map[string]DataFile {
	out := map[string]DataFile{}
	out["file"] = ws.File
	return out
}

func (ws *WorkflowFileCheck) GetName() string {
	return ws.File.Abs()
}

func (ws *WorkflowFileCheck) GetDesc() string {
	return fmt.Sprintf("check-file: %s", ws.File.Abs())
}
