package keyauth

import (
	"context"
	"errors"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/myhaiting/go-fly-lib/satoken"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ErrMissingOrMalformedAPIKey When there is no request of the key thrown ErrMissingOrMalformedAPIKey
var ErrMissingOrMalformedAPIKey = errors.New("missing or malformed API Key")

// Option is the only struct that can be used to set Options.
type Option struct {
	F func(o *Options)
}

type KeyAuthFilterHandler func(c context.Context, ctx *app.RequestContext) bool

type KeyAuthVerifyHandler func(c context.Context, ctx *app.RequestContext) error

type KeyAuthErrorHandler func(context.Context, *app.RequestContext, error)

type Options struct {
	// filterHandler defines a function to skip middleware.
	// Optional. Default: nil
	filterHandler KeyAuthFilterHandler

	verifyHandler KeyAuthVerifyHandler

	// errorHandler defines a function which is executed for an invalid key.
	// It may be used to define a custom error.
	// Optional. Default: 401 Invalid or expired key
	errorHandler KeyAuthErrorHandler

	// keyLookup is a string in the form of "<source>:<name>" that is used
	// to extract key from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "form:<name>"
	// - "param:<name>"
	// - "cookie:<name>"
	keyLookup string

	// authScheme to be used in the Authorization header.
	// Optional. Default value "Bearer".
	authScheme string

	// Manager
	mgr *satoken.Manager
}

func (o *Options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		errorHandler: func(c context.Context, ctx *app.RequestContext, err error) {
			// 如果是没有Token参数报400
			if errors.Is(err, ErrMissingOrMalformedAPIKey) {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, utils.H{
					"code": 1,
					"msg":  err.Error(),
				})
				return
			}
			// Token不可用
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.H{
				"code": 1,
				"msg":  err.Error(),
			})
		},
		authScheme: "Bearer",
		keyLookup:  "header:" + consts.HeaderAuthorization,
	}
	options.Apply(opts)
	if options.mgr == nil {
		panic("satoken manager not found")
	}
	return options
}

func WithFilter(f KeyAuthFilterHandler) Option {
	return Option{
		F: func(o *Options) {
			o.filterHandler = f
		},
	}
}

func WithVerify(f KeyAuthVerifyHandler) Option {
	return Option{
		F: func(o *Options) {
			o.verifyHandler = f
		},
	}
}

func WithErrorHandler(f KeyAuthErrorHandler) Option {
	return Option{
		F: func(o *Options) {
			o.errorHandler = f
		},
	}
}

func WithManager(f *satoken.Manager) Option {
	return Option{
		F: func(o *Options) {
			o.mgr = f
		},
	}
}

func WithKeyLookUp(lookup, authScheme string) Option {
	return Option{func(o *Options) {
		o.keyLookup = lookup
		o.authScheme = authScheme
	}}
}
