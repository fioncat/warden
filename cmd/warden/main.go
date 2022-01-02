package main

import (
	"io/ioutil"

	"github.com/fioncat/warden/config"
	"github.com/fioncat/warden/pkg/debug"
	"github.com/fioncat/warden/pkg/job"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var cmd = &cobra.Command{
	Use: "warden",

	Run: func(cmd *cobra.Command, args []string) {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			debug.Fatal(err, "Failed to read the config")
		}

		jobs := make(map[string]*config.Job)
		err = yaml.Unmarshal(data, &jobs)
		if err != nil {
			debug.Fatal(err, "Failed to parse yaml")
		}

		jobCfg := jobs[jobName]
		if jobCfg == nil {
			debug.Fatalf("Can't find the job %s", jobName)
		}
		err = jobCfg.Normalize()
		if err != nil {
			debug.Fatal(err, "Failed to normalize config")
		}

		job, err := job.New(jobCfg, args)
		if err != nil {
			debug.Fatal(err, "Failed to init the job")
		}
		job.Run()
	},
}

var (
	jobName string
	path    string
)

func init() {
	cmd.PersistentFlags().StringVarP(&path, "file", "f", ".warden.yaml", "the yaml file path")
	cmd.PersistentFlags().StringVarP(&jobName, "job", "j", "main", "job name to execute")
	cmd.PersistentFlags().BoolVarP(&debug.Enable, "debug", "", false, "enable debug mode, this will show extract logs")
}

func main() {
	cmd.Execute()
}
