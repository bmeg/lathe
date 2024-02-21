package scriptfile

import (
	"fmt"

	"github.com/bmeg/lathe/logger"
	"github.com/dop251/goja"
)

type WorkflowDesc struct {
	Name  string
	Steps []Step
}

//func (pd *ProcessDesc) Depends(p *ProcessDesc) {
//fmt.Printf("Adding process dependency: %#v", pd)
//	pd.Dependencies = append(pd.Dependencies, p)
//}

func (wd *WorkflowDesc) Add(call goja.ConstructorCall) *goja.Object {
	if len(call.Arguments) != 1 {
		return nil
	}
	//fmt.Printf("Adding %#v\n", call.Arguments[0])
	e := call.Arguments[0].Export()
	if proc, ok := e.(*ProcessDesc); ok {
		if proc.Name == "" {
			proc.Name = fmt.Sprintf("%s:%d", wd.Name, len(wd.Steps))
		}
		wd.Steps = append(wd.Steps, proc)
	} else if wf, ok := e.(*WorkflowDesc); ok {
		wd.Steps = append(wd.Steps, wf.Steps...)
	} else if file, ok := e.(*File); ok {
		logger.Debug("Adding file check", "path", file)
		wd.Steps = append(wd.Steps, &FileCheck{File: file})
	} else {
		logger.Error("Unknown object", "error", e)
	}
	return nil
}
