package debug

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
)

var Enable bool

// 2021-01-01 15:32:12 [warden] [debug]

var (
	debugLogger *log.Logger
	errLogger   *log.Logger
	execLogger  *log.Logger
	fatalLogger *log.Logger
)

const logFlags = log.Lmsgprefix | log.Ldate | log.Ltime

func init() {
	debugLogger = log.New(os.Stdout, newPrefix(color.CyanString("[debug]")), logFlags)
	errLogger = log.New(os.Stderr, newPrefix(color.RedString("[error]")), logFlags)
	execLogger = log.New(os.Stdout, newPrefix(color.GreenString("[exec]")), logFlags)
	fatalLogger = log.New(os.Stderr, newPrefix(color.RedString("[fatal]")), logFlags)
}

func newPrefix(s string) string {
	return fmt.Sprintf("[warden] %s ", s)
}

func Info(msg ...interface{}) {
	if !Enable {
		return
	}
	debugLogger.Println(msg...)
}

func Infof(msg string, vs ...interface{}) {
	if !Enable {
		return
	}
	debugLogger.Printf(msg, vs...)
}

func Exec(v ...interface{}) {
	execLogger.Println(v...)
}

func Error(err error, msg string, v ...interface{}) {
	msg = fmt.Sprintf(msg, v...)
	errLogger.Printf("%s: %v", msg, err)
}

func Errorf(msg string, v ...interface{}) {
	errLogger.Printf(msg, v...)
}

func Fatal(err error, msg string) {
	fatalLogger.Printf("%s: %v", msg, err)
	os.Exit(1)
}
