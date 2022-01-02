package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fioncat/warden/config"
	"github.com/fioncat/warden/pkg/debug"
	"github.com/fioncat/warden/pkg/watch"
)

func main() {
	var path []string
	var ignores []string
	switch len(os.Args) {
	case 2:
		path = strings.Split(os.Args[1], ",")

	case 3:
		path = strings.Split(os.Args[1], ",")
		ignores = strings.Split(os.Args[2], ",")

	default:
		fmt.Println("Usage: debug-watch <path> [ignores]")
		os.Exit(1)
	}
	debug.Enable = true

	watcher, err := watch.Run(&config.Watch{
		Ignore:  ignores,
		Pattern: path,
	})
	if err != nil {
		debug.Fatal(err, "Failed to run the watcher")
	}

	for range watcher.Notify() {
		fmt.Println("File changed !!!!!! Go to rebuild!!!!!")
	}
}
