package workflow

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/shlex"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

var PoolErrorNotAvailable = Error("not avalible")

type CommandLineTool struct {
	CommandLine string
	BaseDir     string
	Inputs      []string
	Outputs     []string
	NCpus       uint
	MemMB       uint
}

type CommandLog struct {
}

type CommandRunner interface {
	RunCommand(*CommandLineTool) (*CommandLog, error)
}

type SingleMachineRunner struct {
	MaxCPUs  uint
	MaxMemMB uint
	resMutex *sync.Mutex
	memPool  *ConstraintPool
	cpuPool  *ConstraintPool
}

func NewSingleMachineRunner(ncpus uint, maxmb uint) CommandRunner {
	return &SingleMachineRunner{
		MaxCPUs:  ncpus,
		MaxMemMB: maxmb,
		resMutex: &sync.Mutex{},
		memPool:  NewConstraintPool(uint(maxmb)),
		cpuPool:  NewConstraintPool(uint(ncpus)),
	}
}

func (sc *SingleMachineRunner) RunCommand(cmdTool *CommandLineTool) (*CommandLog, error) {
	workdir, _ := filepath.Abs(cmdTool.BaseDir)

	cmdLine, err := shlex.Split(cmdTool.CommandLine)
	if err != nil {
		return nil, err
	}

	var cpuAlloc *PoolAllocation
	var memAlloc *PoolAllocation

	fmt.Printf("Allocating CPU: %d RAM: %d\n", cmdTool.NCpus, cmdTool.MemMB)
	for {
		loopMutex := &sync.Mutex{}
		cpuAlloc, err = sc.cpuPool.Allocate(cmdTool.NCpus)
		if err == nil {
			memAlloc, err = sc.memPool.Allocate(cmdTool.MemMB)
			if err == nil {
				break
			} else {
				cpuAlloc.Return()
				loopMutex.Lock()
				sc.memPool.AddCallback(func() { loopMutex.Unlock() })
			}
		} else {
			loopMutex.Lock()
			sc.cpuPool.AddCallback(func() { loopMutex.Unlock() })
		}
		loopMutex.Lock()
	}
	/*
		sc.memPool.mutext.Lock()
		fmt.Printf("Ram alloced total: %d %#v\n", len(sc.memPool.allocations), sc.memPool.allocations)
		for _, d := range sc.memPool.allocations {
			fmt.Printf("\t%d\n", d.size)
		}
		sc.memPool.mutext.Unlock()
	*/
	defer cpuAlloc.Return()
	defer memAlloc.Return()
	cmd := exec.Command(cmdLine[0], cmdLine[1:]...)
	cmd.Dir = workdir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	log.Printf("(%s) %s %s", cmd.Dir, cmd.Path, strings.Join(cmd.Args, " "))
	//time.Sleep(5 * time.Second)
	return &CommandLog{}, cmd.Run()
}

type PoolAllocation struct {
	size uint
	id   uint
	pool *ConstraintPool
}

func (pa *PoolAllocation) Return() {
	pa.pool.returnAlloc(pa.id)
}

type ConstraintPool struct {
	poolSize    uint
	allocations map[uint]PoolAllocation
	mutext      sync.Mutex
	callbacks   []func()
}

func NewConstraintPool(max uint) *ConstraintPool {
	return &ConstraintPool{
		poolSize:    max,
		allocations: map[uint]PoolAllocation{},
		mutext:      sync.Mutex{},
		callbacks:   []func(){},
	}
}

func (cp *ConstraintPool) sumSize() uint {
	out := uint(0)
	for _, v := range cp.allocations {
		out += v.size
	}
	return out
}

func (cp *ConstraintPool) minID() uint {
	out := ^uint(0)
	for _, v := range cp.allocations {
		if v.size < out {
			out = v.id
		}
	}
	return out
}

func (cp *ConstraintPool) maxID() uint {
	out := uint(0)
	for _, v := range cp.allocations {
		if v.size > out {
			out = v.id
		}
	}
	return out
}

func (cp *ConstraintPool) Allocate(val uint) (*PoolAllocation, error) {
	cp.mutext.Lock()
	defer cp.mutext.Unlock()
	if val+cp.sumSize() > cp.poolSize {
		return nil, PoolErrorNotAvailable
	}
	newID := uint(0)
	if len(cp.allocations) != 0 {
		m := cp.minID()
		if m == 0 {
			newID = cp.maxID() + 1
		} else {
			newID = m - 1
		}
	}
	out := PoolAllocation{size: val, id: newID, pool: cp}
	cp.allocations[newID] = out
	return &out, nil
}

func (cp *ConstraintPool) invokeCallBacks() {
	for ; len(cp.callbacks) > 0; cp.callbacks = cp.callbacks[1:] {
		go cp.callbacks[0]()
	}
	cp.callbacks = []func(){}
}

func (cp *ConstraintPool) returnAlloc(id uint) {
	cp.mutext.Lock()
	defer cp.mutext.Unlock()
	delete(cp.allocations, id)
	cp.invokeCallBacks()
}

func (cp *ConstraintPool) AddCallback(f func()) {
	cp.mutext.Lock()
	defer cp.mutext.Unlock()
	if len(cp.allocations) == 0 {
		go f()
	} else {
		cp.callbacks = append(cp.callbacks, f)
	}
}
