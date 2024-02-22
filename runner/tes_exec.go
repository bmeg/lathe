package runner

import (
	"context"
	"log"
	"path/filepath"

	"github.com/ohsu-comp-bio/funnel/tes"
)

type TesRunner struct {
	Client       *tes.Client
	DefaultImage string
}

func NewTesRunner(host string, defaultImage string) CommandRunner {
	client, _ := tes.NewClient(host)
	return &TesRunner{
		Client:       client,
		DefaultImage: defaultImage,
	}
}

func (tr *TesRunner) RunCommand(cmdTool *CommandLineTool) (*CommandLog, error) {
	workdir, _ := filepath.Abs(cmdTool.BaseDir)

	inputs := []*tes.Input{}
	for _, i := range cmdTool.Inputs {
		t := tes.Input{
			Path: i,
		}
		inputs = append(inputs, &t)
	}
	outputs := []*tes.Output{}
	for _, i := range cmdTool.Outputs {
		t := tes.Output{
			Path: i,
		}
		outputs = append(outputs, &t)
	}

	task := tes.Task{
		Executors: []*tes.Executor{
			{
				Image:   tr.DefaultImage,
				Command: cmdTool.CommandLine,
				Workdir: workdir,
			},
		},
		Resources: &tes.Resources{
			CpuCores: uint32(cmdTool.NCpus),
			RamGb:    float64(cmdTool.MemMB) / 1024,
		},
		Inputs:  inputs,
		Outputs: outputs,
	}

	resp, err := tr.Client.CreateTask(context.Background(), &task)
	if err != nil {
		return nil, err
	}

	resp.GetId()

	err = tr.Client.WaitForTask(context.Background(), resp.Id)
	if err != nil {
		log.Printf("Task Error: %s", err)
	}

	return &CommandLog{}, err
}
