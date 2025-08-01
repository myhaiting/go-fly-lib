package satoken

import (
	"context"
	"sync"
	"time"
)

// NewSession create to token model instance
func newSession(sessionId string, store TokenStore) *Session {
	return &Session{
		Id:            sessionId,
		store:         store,
		CreateTime:    time.Now().UnixNano() / 1e6,
		Data:          make(map[string]any),
		TokenSignList: make([]*TokenSign, 0),
	}
}

type LoginModel struct {
	Device        string `json:"device"`
	Timeout       int    `json:"timeout"`
	ActiveTimeout int    `json:"activeTimeout"`
	Token         string `json:"token"`
}

func (m LoginModel) getDeviceOrDefault() string {
	if m.Device == "" {
		return "default-device"
	}
	return m.Device
}

type TokenSign struct {
	Value  string `json:"value"`
	Device string `json:"device"`
	Tag    any    `json:"tag"`
}

type Session struct {
	sync.Mutex
	Id            string         `json:"id"`
	Type          string         `json:"type"`
	LoginType     string         `json:"loginType"`
	LoginId       any            `json:"loginId,omitempty"`
	Token         string         `json:"token"`
	CreateTime    int64          `json:"createTime"`
	Data          map[string]any `json:"data"`
	TokenSignList []*TokenSign   `json:"tokenSignList"`
	store         TokenStore
	ctx           context.Context
}

func (s *Session) Get(key string) any {
	s.Lock()
	defer s.Unlock()
	if s.Data == nil {
		s.Data = make(map[string]any)
	}
	return s.Data[key]
}

func (s *Session) Set(key string, val any) {
	s.Lock()
	defer s.Unlock()
	if s.Data == nil {
		s.Data = make(map[string]any)
	}
	s.Data[key] = val
}

func (s *Session) Delete(key string) {
	s.Lock()
	defer s.Unlock()
	if s.Data == nil {
		s.Data = make(map[string]any)
	}
	delete(s.Data, key)
}

func (s *Session) addTokenSign(sign TokenSign) {
	s.Lock()
	defer s.Unlock()
	_, oldTokenSign := s.getTokenSign(sign.Value)
	if oldTokenSign == nil {
		s.TokenSignList = append(s.TokenSignList, &sign)
	} else {
		oldTokenSign.Value = sign.Value
		oldTokenSign.Device = sign.Device
		oldTokenSign.Tag = sign.Tag
	}
}

func (s *Session) removeTokenSign(tokenValue string) {
	s.Lock()
	defer s.Unlock()
	index, _ := s.getTokenSign(tokenValue)
	if index >= 0 {
		s.TokenSignList = append(s.TokenSignList[:index], s.TokenSignList[index+1:]...)
	}
}

func (s *Session) getTokenSign(tokenValue string) (int, *TokenSign) {
	if s.TokenSignList == nil {
		return -1, nil
	}
	for index, item := range s.TokenSignList {
		if item.Value == tokenValue {
			return index, item
		}
	}
	return -1, nil
}

func (s *Session) getTokenSignListByDevice(device string) []*TokenSign {
	s.Lock()
	defer s.Unlock()
	ret := make([]*TokenSign, 0)
	for _, item := range s.TokenSignList {
		if device == "" || item.Device == device {
			ret = append(ret, item)
		}
	}
	return ret
}

func (s *Session) getTokenValueListByDevice(device string) ([]string, error) {
	s.Lock()
	defer s.Unlock()
	ret := make([]string, 0)
	for _, item := range s.TokenSignList {
		if item.Device == device {
			ret = append(ret, item.Value)
		}
	}
	return ret, nil
}

func (s *Session) Save() error {
	return s.store.UpdateObj(s.ctx, s.Id, s)
}
