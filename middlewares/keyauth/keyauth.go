package keyauth

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/myhaiting/go-fly-lib/satoken"
	"net/http"
)

const hertzAuthKey = "hertzAuth"

type ctxStore struct {
	Instance   *satoken.Manager // Token Manager
	TokenValue string           // Token value
	LoginId    string           // LoginId
}

func New(opts ...Option) app.HandlerFunc {
	cfg := NewOptions(opts...)
	// token look up
	extractor := LookupTokenFunc(cfg)

	// Return middleware handler
	return func(c context.Context, ctx *app.RequestContext) {
		// Filter request to skip middleware
		if cfg.filterHandler != nil && cfg.filterHandler(c, ctx) {
			ctx.Next(c)
			return
		}
		// Extract and verify key
		tokenValue, err := extractor(ctx)
		if err != nil {
			cfg.errorHandler(c, ctx, err)
			return
		}
		// Get login info
		loginId, err := cfg.mgr.GetLoginId(c, tokenValue)
		if err != nil {
			cfg.errorHandler(c, ctx, err)
			return
		}
		store := &ctxStore{
			Instance:   cfg.mgr,
			TokenValue: tokenValue,
			LoginId:    loginId,
		}
		withValueCtx := context.WithValue(c, hertzAuthKey, store)
		// 参数验证
		if cfg.verifyHandler != nil {
			if err = cfg.verifyHandler(withValueCtx, ctx); err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, utils.H{
					"code": 1,
					"msg":  err.Error(),
				})
				return
			}
		}
		ctx.Next(withValueCtx)
	}
}

func ctxGet(ctx context.Context) (*ctxStore, error) {
	value := ctx.Value(hertzAuthKey)
	if value == nil {
		return nil, fmt.Errorf("auth error: %v", "Config is nil")
	}
	h, ok := value.(*ctxStore)
	if !ok {
		return nil, fmt.Errorf("auth error: %v", "Config is not *Config type")
	}
	return h, nil
}

// Logout logout 当前账户
func Logout(ctx context.Context) error {
	store, err := ctxGet(ctx)
	if err != nil {
		return err
	}
	return store.Instance.LogoutByToken(ctx, store.TokenValue)
}

// LogoutByLoginId 指定用户踢出
func LogoutByLoginId(ctx context.Context, loginId any, device string) error {
	store, err := ctxGet(ctx)
	if err != nil {
		return err
	}
	return store.Instance.LogoutByLoginId(ctx, loginId, device)
}

// GetLoginId get login id
func GetLoginId(ctx context.Context) (string, error) {
	store, err := ctxGet(ctx)
	if err != nil {
		return "", err
	}
	return store.LoginId, nil
}

// GetSession get token session
func GetSession(ctx context.Context) (*satoken.Session, error) {
	store, err := ctxGet(ctx)
	if err != nil {
		return nil, err
	}
	return store.Instance.GetSession(ctx, store.TokenValue, true)
}
