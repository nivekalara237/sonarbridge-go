package cli

import "fmt"

type CliContext struct {
	ConfigFile      string
	Verbose         bool
	LogLevel        string
	serverCert      string
	serverKey       string
	serverSharedKey string
	serverUrl       string
	Output          *Output
}

type OutputFormat = string

const (
	JSON  OutputFormat = "json"
	TABLE OutputFormat = "table"
	TEXT  OutputFormat = "text"
)

var (
	validOutputs = map[string]struct{}{
		"text":  {},
		"json":  {},
		"table": {},
	}

	validLogLevels = map[string]struct{}{
		"debug": {},
		"info":  {},
		"warn":  {},
		"error": {},
	}
)

func (ctx *CliContext) Validate() error {
	if _, ok := validOutputs[ctx.Output.format]; !ok {
		return fmt.Errorf("invalid output format %q (allowed: text, json, table)", ctx.Output.format)
	}
	if _, ok := validLogLevels[ctx.LogLevel]; !ok {
		return fmt.Errorf("invalid log level %q (allowed: debug, info, warn, error)", ctx.LogLevel)
	}
	return nil
}
