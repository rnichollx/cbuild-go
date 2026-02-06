package ccommon

import "context"

type applyProjectDefaultsCtxKeyType struct{}

var applyProjectDefaultsCtxKey = applyProjectDefaultsCtxKeyType{}

func WithApplyProjectDefaults(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, applyProjectDefaultsCtxKey, enabled)
}

func applyProjectDefaultsEnabled(ctx context.Context) bool {
	enabled, ok := ctx.Value(applyProjectDefaultsCtxKey).(bool)
	return ok && enabled
}
