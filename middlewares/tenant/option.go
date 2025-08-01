package tenant

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Option is the only struct that can be used to set Options.
type Option struct {
	F func(o *Options)
}

type FilterHandler func(c context.Context, ctx *app.RequestContext) bool
type EmptyHandler func(context.Context, *app.RequestContext)
type ValueHandler func(context.Context, *app.RequestContext)

type Options struct {
	// tenantKey tenant key
	tenantKey string
	// filterHandler defines a function to skip middleware.
	// Optional. Default: nil
	filterHandler FilterHandler
	emptyHandler  EmptyHandler
	valueHandler  ValueHandler
}

func (o *Options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		tenantKey: "x-t-id",
		filterHandler: func(ctx context.Context, c *app.RequestContext) bool {
			return false
		},
		emptyHandler: func(ctx context.Context, c *app.RequestContext) {
			c.AbortWithStatusJSON(consts.StatusBadRequest, utils.H{
				"code": 1,
				"msg":  "tenant not found",
			})
		},
		valueHandler: func(ctx context.Context, c *app.RequestContext) {
			c.Next(ctx)
		},
	}
	options.Apply(opts)
	return options
}

func WithTenantKey(tenantKey string) Option {
	return Option{func(o *Options) {
		o.tenantKey = tenantKey
	}}
}

func WithFilter(f FilterHandler) Option {
	return Option{
		F: func(o *Options) {
			o.filterHandler = f
		},
	}
}

func WithEmpty(f EmptyHandler) Option {
	return Option{
		F: func(o *Options) {
			o.emptyHandler = f
		},
	}
}

func WithValue(f ValueHandler) Option {
	return Option{
		F: func(o *Options) {
			o.valueHandler = f
		},
	}
}
