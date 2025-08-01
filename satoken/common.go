package satoken

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"strings"
)

const (
	NOT_TOKEN     = "-1"
	INVALID_TOKEN = "-2"
	TOKEN_TIMEOUT = "-3"
	BE_REPLACED   = "-4"
	KICK_OUT      = "-5"
	TOKEN_FREEZE  = "-6"
	NO_PREFIX     = "-7"
)

func (m *Manager) getConfigOrGlobal() *Config {
	return m.cfg
}

func (m *Manager) splicingKeySession(loginId any) string {
	return fmt.Sprintf("%s:%s:session:%s", m.getConfigOrGlobal().TokenName, m.loginType, cast.ToString(loginId))
}

func (m *Manager) splicingKeyTokenValue(tokenValue string) string {
	return fmt.Sprintf("%s:%s:token:%s", m.getConfigOrGlobal().TokenName, m.loginType, tokenValue)
}

func (m *Manager) splicingKeyTokenSession(tokenValue string) string {
	return fmt.Sprintf("%s:%s:token-session:%s", m.getConfigOrGlobal().TokenName, m.loginType, tokenValue)
}

func (m *Manager) splicingKeyLastActiveTime(tokenValue string) string {
	return fmt.Sprintf("%s:%s:last-active:%s", m.getConfigOrGlobal().TokenName, m.loginType, tokenValue)
}

func (m *Manager) splicingKeyDisable(loginId any, service string) string {
	return fmt.Sprintf("%s:%s:disable:%s:%s", m.getConfigOrGlobal().TokenName, m.loginType, service, cast.ToString(loginId))
}

func (m *Manager) splicingKeySafe(tokenValue string, service string) string {
	return fmt.Sprintf("%s:%s:safe:%s:%s", m.getConfigOrGlobal().TokenName, m.loginType, service, tokenValue)
}

func (m *Manager) getSession(ctx context.Context, sessionId string) (*Session, error) {
	var sess Session
	if err := m.tokenStore.GetObj(ctx, sessionId, &sess); err != nil {
		return nil, err
	}
	sess.store = m.tokenStore
	sess.ctx = ctx
	return &sess, nil
}

func (m *Manager) deleteSession(ctx context.Context, sessionId string) error {
	return m.tokenStore.DeleteObj(ctx, sessionId)
}

func (m *Manager) createLoginSession(ctx context.Context, loginId any, model LoginModel) (string, error) {
	tokenValue, err := m.genTokenValue(ctx, loginId, model)
	if err != nil {
		return "", err
	}
	sess, err := m.getSessionByLoginId(ctx, loginId, true)
	if err != nil {
		return "", err
	}
	sess.addTokenSign(TokenSign{
		Value:  tokenValue,
		Device: model.getDeviceOrDefault(),
		Tag:    "",
	})
	if err = sess.Save(); err != nil {
		return "", err
	}
	if err = m.tokenStore.Set(ctx, m.splicingKeyTokenValue(tokenValue), cast.ToString(loginId), m.getConfigOrGlobal().Timeout); err != nil {
		return "", err
	}
	return tokenValue, nil
}

func (m *Manager) setTokenValue(ctx context.Context, tokenValue string) error {
	return nil
}

func (m *Manager) genTokenValue(ctx context.Context, loginId any, model LoginModel) (string, error) {

	// 顶人下线
	if !m.getConfigOrGlobal().IsConcurrent {
		if err := m.replaced(ctx, loginId, model.Device); err != nil {
			return "", err
		}
	}

	// 自定义Token使用
	if model.Token != "" {
		return model.Token, nil
	}

	// Token复用
	if m.getConfigOrGlobal().IsConcurrent {
		if m.getConfigOrGlobal().IsShare {
			tokenValue, err := m.getTokenValueByLoginId(ctx, loginId, model.getDeviceOrDefault())
			if err != nil {
				return "", err
			}
			if tokenValue != "" {
				return tokenValue, nil
			}
		}
	}
	return m.createTokenValue()
}

// 顶人下线，根据账号id 和 设备类型
func (m *Manager) replaced(ctx context.Context, loginId any, device string) error {
	sess, err := m.getSessionByLoginId(ctx, loginId, false)
	if err != nil {
		if errors.Is(err, ErrObjectNotExist) {
			return nil
		}
		return err
	}

	signList := sess.getTokenSignListByDevice(device)

	for _, sign := range signList {
		sess.removeTokenSign(sign.Value)
		// save session
		if err = sess.Save(); err != nil {
			return err
		}
		// update token mapping
		if err = m.tokenStore.Update(ctx, m.splicingKeyTokenValue(sign.Value), BE_REPLACED); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) getTokenValueByLoginId(ctx context.Context, loginId any, device string) (string, error) {
	values, err := m.getTokenValueListByLoginId(ctx, loginId, device)
	if err != nil {
		return "", err
	}
	if len(values) == 0 {
		return "", nil
	}
	return values[0], nil
}

func (m *Manager) getTokenValueListByLoginId(ctx context.Context, loginId any, device string) ([]string, error) {
	sess, err := m.getSessionByLoginId(ctx, loginId, false)
	if err != nil {
		if errors.Is(err, ErrObjectNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	return sess.getTokenValueListByDevice(device)
}

func (m *Manager) deleteTokenToIdMapping(ctx context.Context, tokenValue string) error {
	return m.tokenStore.Delete(ctx, m.splicingKeyTokenValue(tokenValue))
}

func (m *Manager) deleteTokenSession(ctx context.Context, tokenValue string) error {
	return m.tokenStore.Delete(ctx, m.splicingKeyTokenSession(tokenValue))
}

func (m *Manager) getLoginIdNotHandle(ctx context.Context, tokenValue string) string {
	loginId, err := m.tokenStore.Get(ctx, m.splicingKeyTokenValue(tokenValue))
	if err != nil {
		return ""
	}
	return loginId
}

func (m *Manager) getSessionByLoginId(ctx context.Context, loginId any, isCreate bool) (*Session, error) {
	return m.getSessionBySessionId(ctx, m.splicingKeySession(loginId), isCreate, func(sess *Session) {
		sess.Type = SessionTypeAccount
		sess.LoginType = m.loginType
		sess.LoginId = loginId
	})
}

func (m *Manager) getTokenSessionByToken(ctx context.Context, tokenValue string, isCreate bool) (*Session, error) {
	return m.getSessionBySessionId(ctx, m.splicingKeyTokenSession(tokenValue), isCreate, func(sess *Session) {
		sess.Type = SessionTypeToken
		sess.LoginType = m.loginType
		sess.Token = tokenValue
	})
}

func (m *Manager) getSessionBySessionId(ctx context.Context, sessionId string, isCreate bool, callback func(*Session)) (*Session, error) {
	if sessionId == "" {
		return nil, fmt.Errorf("session id is empty")
	}
	sess, err := m.getSession(ctx, sessionId)
	if err != nil {
		if errors.Is(err, ErrObjectNotExist) && isCreate {
			sess = newSession(sessionId, m.tokenStore)
			sess.ctx = ctx
			if callback != nil {
				callback(sess)
			}
			if err = m.tokenStore.SetObj(ctx, sessionId, sess, m.getConfigOrGlobal().Timeout); err != nil {
				return nil, err
			}
			return sess, nil
		}
		return nil, err
	}
	return sess, nil
}

func (m *Manager) isValidLoginId(loginId string) error {
	switch loginId {
	case NOT_TOKEN, "":
		return ErrNoToken
	case INVALID_TOKEN:
		return ErrInvalidToken
	case TOKEN_TIMEOUT:
		return ErrTokenTimeout
	case BE_REPLACED:
		return ErrBeReplaced
	case KICK_OUT:
		return ErrKickOut
	case TOKEN_FREEZE:
		return ErrTokenFreeze
	case NO_PREFIX:
		return ErrNoPrefix
	}
	return nil
}

// 生成Token值
func (m *Manager) createTokenValue() (string, error) {
	return strings.ReplaceAll(uuid.New().String(), "-", ""), nil
}
