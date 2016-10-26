package lib

import "fmt"

type Logger interface {
	Info(string, ...interface{})
	Debug(string, ...interface{})
}

func CreateLogger(verbose bool, silent bool) Logger {
	return &stdoutLogger{
		verbose: verbose,
		silent:  silent,
	}
}

type stdoutLogger struct {
	verbose bool // If !verbose, suppress Debug messages
	silent  bool // If silent, suppress Info messages.  Takes priority.
}

func (lg stdoutLogger) Info(format string, args ...interface{}) {
	if !lg.silent {
		fmt.Printf(format, args...)
	}
}

func (lg stdoutLogger) Debug(format string, args ...interface{}) {
	if !lg.silent && lg.verbose {
		fmt.Printf(format, args...)
	}
}
