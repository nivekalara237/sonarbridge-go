package cli

import "fmt"

type Context struct {
	ConfigFile string
	Verbose    bool
	LogLevel   string
	Output     *Output
}

type OutputFormat = string

const (
	JSON  = "json"
	TABLE = "table"
	TEXT  = "text"
)

var (
	validOutputs = map[string]struct{}{
		"text": {},
		"json": {},
	}

	validLogLevels = map[string]struct{}{
		"debug": {},
		"info":  {},
		"warn":  {},
		"error": {},
	}
)

func (ctx *Context) Validate() error {
	if _, ok := validOutputs[ctx.Output.format]; !ok {
		return fmt.Errorf("invalid output format %q (allowed: text, json, table)", ctx.Output.format)
	}
	if _, ok := validLogLevels[ctx.LogLevel]; !ok {
		return fmt.Errorf("invalid log level %q (allowed: debug, info, warn, error)", ctx.LogLevel)
	}
	return nil
}
