package main

import (
	"io/ioutil"

	"github.com/fioncat/warden/config"
	"github.com/fioncat/warden/pkg/debug"
	"github.com/fioncat/warden/pkg/job"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const usage = `Warden is a simple command line tool to help you reload the program
after code is changed.

This command should be executed in the root path of the project.

The file '.warden.yaml' need to be created first to configure the job.
You can also use flag '-f <filename>' to specify the config file path.

The definition of the file can be found in the README:
  https://github.com/fioncat/warden/README.md

You can use flag '-n <job-name>' to specify the job to execute, the
default is 'main'.

The args of this command will be passed to exec stage in the job directly,
you can use this feature to pass extra args to the program.

For more usage, please refer to the README in github.`

var cmd = &cobra.Command{
	Use: "warden",

	Long: usage,

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
	cmd.PersistentFlags().StringVarP(&jobName, "name", "n", "main", "job name to execute")
	cmd.PersistentFlags().BoolVarP(&debug.Enable, "debug", "", false, "enable debug mode, this will show extract logs")
}

func main() {
	cmd.Execute()
}
