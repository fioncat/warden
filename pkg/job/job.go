package job

import (
	"fmt"
	"time"

	"github.com/fioncat/warden/config"
	"github.com/fioncat/warden/pkg/debug"
	"github.com/fioncat/warden/pkg/exec"
	"github.com/fioncat/warden/pkg/watch"
)

type Job struct {
	watcher *watch.Watcher

	builds []*exec.Cmd

	process *exec.Process

	kill chan struct{}

	delay time.Duration
}

func New(cfg *config.Job, args []string) (*Job, error) {
	job := new(Job)
	var err error

	job.watcher, err = watch.Run(cfg.Watch)
	if err != nil {
		return nil, err
	}

	job.builds = make([]*exec.Cmd, len(cfg.Build))
	for i, build := range cfg.Build {
		if len(cfg.Env) > 0 {
			build.Env = mergeEnv(cfg.Env, build.Env)
		}
		cmd, err := exec.Command(build, nil)
		if err != nil {
			return nil, fmt.Errorf("init job: failed "+
				"to init build command %s: %v",
				build.Script, err)
		}
		job.builds[i] = cmd
	}

	if cfg.Exec != nil {
		if len(cfg.Env) > 0 {
			cfg.Exec.Env = mergeEnv(cfg.Env, cfg.Exec.Env)
		}
		job.process, err = exec.NewProcess(cfg.Exec, args)
		if err != nil {
			return nil, fmt.Errorf("init job: failed to "+
				"init exec command %s: %v", cfg.Exec.Script, err)
		}
	} else {
		debug.Info("Can't find exec for current job, use default sleep exec")
		job.process, _ = exec.NewProcess(&config.Exec{
			Script: "sleep 3600",
		}, nil)
	}

	job.kill = make(chan struct{})

	if cfg.Delay != "" {
		job.delay, err = time.ParseDuration(cfg.Delay)
		if err != nil {
			return nil, fmt.Errorf("init job: parse delay %q failed: %v",
				cfg.Delay, err)
		}
		if job.delay < time.Second {
			return nil, fmt.Errorf("init job: delay is too small, " +
				"the minimum value is '1s'")
		}
		if job.delay > time.Minute {
			return nil, fmt.Errorf("init job: delay is too big, " +
				"the maximum value is '60s'")
		}
	} else {
		debug.Info("Use default delay 3s for current job")
		job.delay = time.Second * 3
	}

	return job, nil
}

func (job *Job) Run() {
	go job.watchChange()
	for {
		ok, err := job.exec()
		if err != nil {
			debug.Error(err, "Process exited with error")
		}
		if ok {
			continue
		}
		_, ok = <-job.kill
		if !ok {
			debug.Info("Job exiting")
			return
		}
	}
}

func (job *Job) exec() (bool, error) {
	for _, build := range job.builds {
		err := build.Run()
		if err != nil {
			return false, err
		}
	}

	done := make(chan error, 1)
	go func() {
		err := job.process.Run()
		done <- err
		debug.Infof("Job: The process exit with: %v", err)
	}()

	select {
	case _, ok := <-job.kill:
		job.process.Kill()
		if !ok {
			debug.Info("Job: The process was interrupted by closing")
		}
		debug.Info("Job: The process was killed")
		return true, nil

	case err := <-done:
		return false, err
	}
}

func (job *Job) watchChange() {
	notify := job.watcher.Notify()
	for e := range notify {
		timer := time.NewTimer(job.delay)
		debug.Infof("Job: Received change: %s", e.Name)
	waitLoop:
		for {
			select {
			case <-timer.C:
				break waitLoop

			case e := <-notify:
				debug.Infof("Job: Discarded overflow change: %s", e.Name)
			}
		}
		job.kill <- struct{}{}
	}
}

func mergeEnv(a, b config.Env) config.Env {
	r := make(config.Env, len(a)+len(b))
	for key, val := range a {
		r[key] = val
	}
	for key, val := range b {
		r[key] = val
	}
	return r
}
