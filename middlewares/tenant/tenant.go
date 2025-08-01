package tenant

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/savsgio/gotils/strconv"
	"github.com/spf13/cast"
)

const hertzTenantKey = "hertzTenantKey"

func New(opts ...Option) app.HandlerFunc {
	cfg := NewOptions(opts...)
	return func(c context.Context, ctx *app.RequestContext) {
		// Filter request to skip middleware
		if cfg.filterHandler != nil && cfg.filterHandler(c, ctx) {
			ctx.Next(c)
			return
		}
		// Get Value from header
		tenantValue := strconv.B2S(ctx.GetHeader(cfg.tenantKey))
		if tenantValue == "" {
			cfg.emptyHandler(c, ctx)
			return
		}
		withValueCtx := context.WithValue(c, hertzTenantKey, tenantValue)
		cfg.valueHandler(withValueCtx, ctx)
	}
}

// Get get tenant value
func Get(ctx context.Context) string {
	return cast.ToString(ctx.Value(hertzTenantKey))
}
