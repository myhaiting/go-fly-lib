package satoken

import (
	"context"
	"time"
)

// TokenStore the token information storage interface
type TokenStore interface {
	Get(context.Context, string) (string, error)
	Set(context.Context, string, string, time.Duration) error
	Update(context.Context, string, string) error
	Delete(context.Context, string) error
	GetTimeout(context.Context, string) (time.Duration, error)
	UpdateTimeout(context.Context, string, time.Duration) error

	GetObj(context.Context, string, any) error
	SetObj(context.Context, string, any, time.Duration) error
	UpdateObj(context.Context, string, any) error
	DeleteObj(context.Context, string) error
	GetObjTimeout(context.Context, string) (time.Duration, error)
	UpdateObjTimeout(context.Context, string, time.Duration) error
}
