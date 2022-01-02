package exec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fioncat/warden/config"
	"github.com/fioncat/warden/pkg/debug"
)

type Cmd struct {
	name   string
	args   []string
	extEnv []string
}

func Command(cfg *config.Exec, extArgs []string) (*Cmd, error) {
	tmp := strings.Fields(cfg.Script)
	if len(tmp) == 0 {
		return nil, errors.New("script cannot be empty")
	}

	name := tmp[0]
	var args []string
	if len(tmp) > 1 {
		args = tmp[1:]
	}
	if len(extArgs) > 0 {
		args = append(args, extArgs...)
	}

	var extEnv []string
	if len(cfg.Env) > 0 {
		for key, val := range cfg.Env {
			v := fmt.Sprintf("%s=%s", key, val)
			extEnv = append(extEnv, v)
		}
	}

	return &Cmd{
		name:   name,
		args:   args,
		extEnv: extEnv,
	}, nil
}

func (cmd *Cmd) show() {
	if len(cmd.extEnv) > 0 {
		debug.Execf("%s %s %s", strings.Join(cmd.extEnv, " "),
			cmd.name, strings.Join(cmd.args, " "))
	} else {
		debug.Execf("%s %s", cmd.name, strings.Join(cmd.args, " "))
	}
}

func (cmd *Cmd) osCmd() *exec.Cmd {
	osCmd := exec.Command(cmd.name, cmd.args...)
	osCmd.Stdout = os.Stdout
	osCmd.Stderr = os.Stderr
	osCmd.Stdin = os.Stdin
	if len(cmd.extEnv) > 0 {
		osCmd.Env = os.Environ()
		osCmd.Env = append(osCmd.Env, cmd.extEnv...)
	}
	return osCmd
}

func (cmd *Cmd) Run() error {
	cmd.show()
	return cmd.osCmd().Run()
}

func (cmd *Cmd) start() (*os.Process, func() error, error) {
	cmd.show()
	osCmd := cmd.osCmd()
	err := osCmd.Start()
	if err != nil {
		return nil, nil, err
	}
	return osCmd.Process, func() error {
		return osCmd.Wait()
	}, nil
}

type Process struct {
	cmd *Cmd

	process *os.Process

	running uint32
}

func NewProcess(cfg *config.Exec, args []string) (*Process, error) {
	cmd, err := Command(cfg, args)
	if err != nil {
		return nil, err
	}
	return &Process{cmd: cmd}, nil
}

func (p *Process) Run() error {
	atomic.StoreUint32(&p.running, 1)
	process, wait, err := p.cmd.start()
	if err != nil {
		return err
	}
	p.process = process
	err = wait()
	if atomic.LoadUint32(&p.running) == 0 {
		return nil
	}
	atomic.StoreUint32(&p.running, 0)
	return err
}

func (p *Process) Kill() {
	if atomic.LoadUint32(&p.running) == 0 || p.process == nil {
		debug.Info("Kill: the process was not running, skip killing")
		return
	}
	for {
		err := p.process.Kill()
		if err == nil {
			atomic.StoreUint32(&p.running, 0)
			return
		}
		debug.Error(err, "Failed to kill the process %d: %v, we will retry in 1s",
			p.process.Pid, err)
		time.Sleep(time.Second)
	}
}
