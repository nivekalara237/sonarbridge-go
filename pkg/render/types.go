package render

import "sonarbridge-go/pkg/render/loader"

type Request struct {
	Ref    loader.Ref
	Format Format
	Data   any
}
