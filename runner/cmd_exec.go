package runner

import (
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bmeg/lathe/logger"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

var PoolErrorNotAvailable = Error("not avalible")

type CommandLineTool struct {
	CommandLine []string
	BaseDir     string
	Inputs      []string
	Outputs     []string
	NCpus       uint
	MemMB       uint
	Image       string
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

	var err error
	var cpuAlloc *PoolAllocation
	var memAlloc *PoolAllocation

	logger.Info("ResourceRequest", "cpus", cmdTool.NCpus, "memMB", cmdTool.MemMB)
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
		log.Printf("Ram alloced total: %d %#v\n", len(sc.memPool.allocations), sc.memPool.allocations)
		for _, d := range sc.memPool.allocations {
			log.Printf("\t%d\n", d.size)
		}
		sc.memPool.mutext.Unlock()
	*/
	defer cpuAlloc.Return()
	defer memAlloc.Return()
	var cmd *exec.Cmd
	if cmdTool.Image != "" {
		dockerCmd := []string{"docker", "run", "--rm"}
		u, _ := user.Current()
		dockerCmd = append(dockerCmd, "--user", u.Uid)
		dockerCmd = append(dockerCmd, "-v", workdir+":"+workdir)
		dockerCmd = append(dockerCmd, "-w", workdir)

		for _, i := range cmdTool.Inputs {
			p, _ := filepath.Abs(filepath.Join(workdir, i))
			dockerCmd = append(dockerCmd, "-v", p+":"+p)
		}
		oSet := map[string]bool{}
		for _, i := range cmdTool.Outputs {
			p, _ := filepath.Abs(filepath.Join(workdir, i))
			b := filepath.Dir(p)
			oSet[b] = true
		}
		for b := range oSet {
			dockerCmd = append(dockerCmd, "-v", b+":"+b)
		}
		dockerCmd = append(dockerCmd, cmdTool.Image)
		dockerCmd = append(dockerCmd, cmdTool.CommandLine...)
		logger.Info("Executing", "dockerCommand", strings.Join(dockerCmd, " "))
		cmd = exec.Command(dockerCmd[0], dockerCmd[1:]...)
		cmd.Dir = workdir
	} else {
		logger.Info("Executing", "commandLine", cmdTool.CommandLine)
		cmd = exec.Command(cmdTool.CommandLine[0], cmdTool.CommandLine[1:]...)
		cmd.Dir = workdir
	}
	//TODO: manager tool output capture
	//cmd.Stdout = os.Stderr
	//cmd.Stderr = os.Stderr
	logger.Debug("(%s) %s %s", cmd.Dir, cmd.Path, strings.Join(cmd.Args, " "))
	//time.Sleep(5 * time.Second)
	err = cmd.Run()
	if err != nil {
		logger.Error("Command exited with error", "commandLine", cmdTool.CommandLine, "error", err)
	}
	return &CommandLog{}, err
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
