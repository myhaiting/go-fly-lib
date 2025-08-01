package satoken

import (
	"context"
	"errors"
	"github.com/myhaiting/go-fly-lib/bizerr"
)

var (
	ErrObjectNotExist = errors.New("object not exist")
	ErrTokenNotExist  = errors.New("token not exist")

	ErrNoToken      = bizerr.New(10000, "satoken.token.notExist")
	ErrInvalidToken = bizerr.New(10001, "satoken.token.invalid")
	ErrTokenTimeout = bizerr.New(10002, "satoken.token.timeout")
	ErrBeReplaced   = bizerr.New(10003, "satoken.token.beReplaced")
	ErrKickOut      = bizerr.New(10004, "satoken.token.beKickOut")
	ErrTokenFreeze  = bizerr.New(10005, "satoken.token.freeze")
	ErrNoPrefix     = bizerr.New(10006, "satoken.token.noPrefix")
)

const (
	SessionTypeAccount = "Account-Session"
	SessionTypeToken   = "Token-Session"
	SessionTypeCustom  = "Custom-Session"
)

// NewDefaultManager create to default authorization management instance
func NewDefaultManager() *Manager {
	m := NewManager("login")
	return m
}

// NewManager create to authorization management instance
func NewManager(loginType string) *Manager {
	return &Manager{
		loginType: loginType,
		cfg:       NewDefaultConfig(),
	}
}

// Manager provide authorization management
type Manager struct {
	cfg        *Config
	tokenStore TokenStore
	loginType  string
}

// SetCfg set the authorization code grant token config
func (m *Manager) SetCfg(cfg *Config) {
	m.cfg = cfg
}

// MapTokenStorage mapping the token store interface
func (m *Manager) MapTokenStorage(store TokenStore) {
	m.tokenStore = store
}

// MustTokenStorage mandatory mapping the token store interface
func (m *Manager) MustTokenStorage(store TokenStore, err error) {
	if err != nil {
		panic(err)
	}
	m.tokenStore = store
}

// GetLoginId getToken
func (m *Manager) GetLoginId(ctx context.Context, token string) (string, error) {
	loginId, err := m.tokenStore.Get(ctx, m.splicingKeyTokenValue(token))
	if err != nil {
		if errors.Is(err, ErrTokenNotExist) {
			return "", bizerr.WrapBizError(ctx, ErrNoToken)
		}
		return "", bizerr.WrapBizError(ctx, err)
	}
	if err = m.isValidLoginId(loginId); err != nil {
		return "", bizerr.WrapBizError(ctx, err)
	}
	return loginId, nil
}

// Login login
func (m *Manager) Login(ctx context.Context, loginId any, model LoginModel) (string, error) {
	tokenValue, err := m.createLoginSession(ctx, loginId, model)
	if err != nil {
		return "", err
	}
	if err = m.setTokenValue(ctx, tokenValue); err != nil {
		return "", err
	}
	return tokenValue, nil
}

// LogoutByLoginId logout
func (m *Manager) LogoutByLoginId(ctx context.Context, loginId any, device string) error {
	sess, err := m.getSessionByLoginId(ctx, loginId, false)
	if err != nil {
		return err
	}
	for _, item := range sess.getTokenSignListByDevice(device) {
		// 删除Token
		sess.removeTokenSign(item.Value)
		// 删除TokenMapping
		if err = m.deleteTokenToIdMapping(ctx, item.Value); err != nil {
			return err
		}
		// 删除Token Session
		if err = m.deleteTokenSession(ctx, item.Value); err != nil {
			return err
		}
	}
	// 如果没有Token则注销会话
	if len(sess.TokenSignList) == 0 {
		// 删除Session
		if err = m.deleteSession(ctx, sess.Id); err != nil {
			return err
		}
	} else {
		// 保存Session
		if err = sess.Save(); err != nil {
			return err
		}
	}
	return nil
}

// LogoutByToken logout
func (m *Manager) LogoutByToken(ctx context.Context, tokenValue string) error {
	// 删除Token Session
	if err := m.deleteTokenSession(ctx, tokenValue); err != nil {
		return err
	}
	// 获取LoginId
	loginId := m.getLoginIdNotHandle(ctx, tokenValue)
	if loginId != "" {
		if err := m.deleteTokenToIdMapping(ctx, tokenValue); err != nil {
			return err
		}
	}
	// 判断Id是否可用
	if err := m.isValidLoginId(loginId); err != nil {
		return nil
	}
	// 获取登录会话
	sess, err := m.getSessionByLoginId(ctx, loginId, false)
	if err != nil {
		if errors.Is(err, ErrObjectNotExist) {
			return nil
		}
		return err
	}
	sess.removeTokenSign(tokenValue)
	// 如果没有Token则注销会话
	if len(sess.TokenSignList) == 0 {
		if err = m.deleteSession(ctx, sess.Id); err != nil {
			return err
		}
	} else {
		if err = sess.Save(); err != nil {
			return err
		}
	}
	return nil
}

// GetSession session
func (m *Manager) GetSession(ctx context.Context, tokenValue string, isCreate bool) (*Session, error) {
	_, err := m.GetLoginId(ctx, tokenValue)
	if err != nil {
		return nil, err
	}
	return m.getTokenSessionByToken(ctx, tokenValue, isCreate)
}
