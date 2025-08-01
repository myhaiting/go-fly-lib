package bizerr

import (
	"context"
	"errors"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/i18n"
	"strings"
)

type bizError struct {
	c int
	s string
}

// New returns an error that formats as the given text.
func New(code int, key string) error {
	return &bizError{code, key}
}

func (e *bizError) Error() string {
	return e.s
}

// WrapBizError 错误信息转换
func WrapBizError(ctx context.Context, err error) error {
	e, ok := err.(*bizError)
	if ok && strings.Contains(e.s, ".") {
		msg, er := i18n.GetMessage(ctx, e.s)
		if er != nil {
			hlog.Warn("i18n: get message error: %v", er)
			return err
		}
		return errors.New(msg)
	}
	return err
}

// BizErrorMsg 获取错误信息
func BizErrorMsg(ctx context.Context, err error) (int, string) {
	e, ok := err.(*bizError)
	if ok {
		if strings.Contains(e.s, ".") {
			msg, er := i18n.GetMessage(ctx, e.s)
			if er != nil {
				hlog.Warn("i18n: get message error: %v", er)
				return e.c, err.Error()
			}
			return e.c, msg
		}
		return e.c, e.s
	}
	return 1, err.Error()
}
