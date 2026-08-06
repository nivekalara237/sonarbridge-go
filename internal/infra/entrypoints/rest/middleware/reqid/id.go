package reqid

import "context"

type ctxKey struct {

}

func With(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func From(ctx context.Context) string {
	id,_ := ctx.Value(ctxKey{}).(string)
	return id
}