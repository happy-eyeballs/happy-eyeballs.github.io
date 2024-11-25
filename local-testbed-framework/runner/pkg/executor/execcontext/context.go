package execcontext

import "context"

const cancelFuncKey = "cancelFunc"

func NewExecContext(parentCtx context.Context) context.Context {
	ctx, cancelCtx := context.WithCancel(parentCtx)
	return context.WithValue(ctx, cancelFuncKey, cancelCtx)
}

func CancelExecContext(ctx context.Context) {
	cancelCtx := ctx.Value(cancelFuncKey)
	if cancelCtx == nil {
		return
	}

	cancelCtxFunc := cancelCtx.(context.CancelFunc)
	cancelCtxFunc()
}
