package im

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/registry/nacos/v2"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/tidwall/gjson"
	"net/http"
)

type Client interface {
	AddUser(uid string, token string, deviceFlag, deviceLevel int) (int, error)
	AddSysUser(uid []string) error
	Route(uid string) (string, string, string, error)
	UserOnlineStatus(uid []string) ([]OnlineStatus, error)
	CreateChannel(channelId string, channelType int, large int, ban int, subscribers []string) error
	RemoveChannel(channelId string, channelType int) error
	SubscriberAdd(channelId string, channelType, reset int, subscribers []string, isTemp int) error
	SubscriberRemove(channelId string, channelType int, subscribers []string) error
	BlacklistAdd(channelId string, channelType int, uids []string) error
	BlacklistRemove(channelId string, channelType int, uids []string) error
	BlacklistSet(channelId string, channelType int, uids []string) error
	WhitelistAdd(channelId string, channelType int, uids []string) error
	WhitelistRemove(channelId string, channelType int, uids []string) error
	WhitelistSet(channelId string, channelType int, uids []string) error
	MessageSend(msg *MsgSendReq) (*MsgSendResp, error)
	ChannelMessageAsync(uid, channelId string, channelType, start, end, limit, pullMode int) (*SyncChannelMessageResp, error)
	QuitUserDevice(uid string, deviceFlag int) error
	GetChannelMaxSeq(channelId string, channelType int) (uint32, error)
	SyncUserConversation(uid string, version int64, msgCount int64, lastMsgSeqs string, larges []*Channel) ([]*Conversation, error)
	ClearConversationUnread(req ClearConversationUnreadReq) error
	DeleteConversation(req DeleteConversationReq) error
	SyncMessage(uid string, messageSeq uint32, limit int) ([]*Message, error)
	SyncMessageAck(uid string, lastMessageSeq uint32) error
	GetWithChannelAndSeqs(channelID string, channelType uint8, loginUID string, seqs []uint32) (*SyncChannelMessageResp, error)
	MessageSearch(req MessageSearchReq) ([]*Message, error)
}

type imClient struct {
	token string
	debug bool
	cli   *client.Client
}

type Result struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

func NewClient(token string, debug bool, namingClient naming_client.INamingClient) Client {
	var err error
	c := &imClient{token: token, debug: debug}
	c.cli, err = client.NewClient()
	if err != nil {
		panic(err)
	}
	var resolver discovery.Resolver
	// 使用默认NacosResolver
	if namingClient == nil {
		resolver, err = nacos.NewDefaultNacosResolver()
		if err != nil {
			panic(err)
		}
	} else {
		resolver = nacos.NewNacosResolver(namingClient)
	}
	c.cli.Use(sd.Discovery(resolver))
	return c
}

func (c *imClient) post(path string, params interface{}) ([]byte, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	req.SetRequestURI("http://wukongim" + path)
	req.SetMethod("POST")
	req.Header.Add("content-type", "application/json")
	if c.token != "" {
		req.Header.Add("Authorization", "Bearer "+c.token)
	}
	req.SetBody(data)
	req.SetOptions(config.WithSD(true))
	err = c.cli.Do(context.Background(), req, resp)
	if err != nil {
		return nil, err
	}
	body, err := resp.BodyE()
	if err != nil {
		return nil, err
	}
	if c.debug {
		hlog.Infof("[Req] %s - %s [Res] %d - %s", path, string(data), resp.StatusCode(), string(body))
	}
	if resp.StatusCode() != http.StatusOK {
		var errRes Result
		if err = json.Unmarshal(body, &errRes); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(errRes.Msg)
	}
	return body, nil
}

// UserOnlineStatus 查询在线用户
func (c *imClient) UserOnlineStatus(uid []string) ([]OnlineStatus, error) {
	rsp, err := c.post("/user/onlinestatus", uid)
	if err != nil {
		return nil, err
	}
	var results []OnlineStatus
	if err = json.Unmarshal(rsp, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// Route 路由获取
func (c *imClient) Route(uid string) (string, string, string, error) {
	statusCode, data, err := c.cli.Get(context.Background(), nil, "http://wukongim/route?uid="+uid, config.WithSD(true))
	if err != nil {
		return "", "", "", err
	}
	if statusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("error code: %d, body: %s", statusCode, data)
	}
	var addr struct {
		WsAddr  string `json:"ws_addr"`
		TcpAddr string `json:"tcp_addr"`
		WssAddr string `json:"wss_addr"`
	}
	if err = json.Unmarshal(data, &addr); err != nil {
		return "", "", "", err
	}
	return addr.WsAddr, addr.WssAddr, addr.TcpAddr, nil
}

// AddUser 添加用户
func (c *imClient) AddUser(uid string, token string, deviceFlag, deviceLevel int) (int, error) {
	params := map[string]interface{}{
		"uid":          uid,
		"token":        token,
		"device_flag":  deviceFlag,
		"device_level": deviceLevel,
	}
	res, err := c.post("/user/token", params)
	if err != nil {
		return 0, err
	}
	var resp struct {
		Status int `json:"status"`
	}
	if err = json.Unmarshal(res, &resp); err != nil {
		return 0, err
	}
	return resp.Status, nil
}

// AddSysUser 添加系统账号
func (c *imClient) AddSysUser(uid []string) error {
	_, err := c.post("/user/systemuids_add", uid)
	if err != nil {
		return err
	}
	return nil
}

// CreateChannel 创建频道(聊天组)
func (c *imClient) CreateChannel(channelId string, channelType int, large int, ban int, subscribers []string) error {
	params := map[string]interface{}{
		"channel_id":   channelId,
		"channel_type": channelType,
		"large":        large,
		"ban":          ban,
		"subscribers":  subscribers,
	}
	_, err := c.post("/channel", params)
	if err != nil {
		return err
	}
	return nil
}

// RemoveChannel 删除频道
func (c *imClient) RemoveChannel(channelId string, channelType int) error {
	params := map[string]interface{}{
		"channel_id":   channelId,
		"channel_type": channelType,
	}
	_, err := c.post("/channel/delete", params)
	if err != nil {
		return err
	}
	return nil
}

// SubscriberAdd 添加订阅者
func (c *imClient) SubscriberAdd(channelId string, channelType, reset int, subscribers []string, isTemp int) error {
	hlog.Infof("SubscriberAdd %s %v", channelId, subscribers)
	params := map[string]interface{}{
		"channel_id":      channelId,
		"channel_type":    channelType,
		"reset":           reset,
		"subscribers":     subscribers,
		"temp_subscriber": isTemp,
	}
	_, err := c.post("/channel/subscriber_add", params)
	if err != nil {
		return err
	}
	hlog.Info("SubscriberAdd success")
	return nil
}

// SubscriberRemove 移除订阅者
func (c *imClient) SubscriberRemove(channelId string, channelType int, subscribers []string) error {
	hlog.Infof("SubscriberRemove %s %v", channelId, subscribers)
	params := map[string]interface{}{
		"channel_id":   channelId,
		"channel_type": channelType,
		"subscribers":  subscribers,
	}
	_, err := c.post("/channel/subscriber_remove", params)
	if err != nil {
		return err
	}
	hlog.Info("SubscriberRemove success")
	return nil
}

// BlacklistAdd 添加黑名单
func (c *imClient) BlacklistAdd(channelId string, channelType int, uids []string) error {
	params := map[string]interface{}{
		"channel_id":   channelId,
		"channel_type": channelType,
		"uids":         uids,
	}
	_, err := c.post("/channel/blacklist_add", params)
	if err != nil {
		return err
	}
	return nil
}

// BlacklistRemove 移除黑名单
func (c *imClient) BlacklistRemove(channelId string, channelType int, uids []string) error {
	params := map[string]interface{}{
		"channel_id":   channelId,
		"channel_type": channelType,
		"uids":         uids,
	}
	_, err := c.post("/channel/blacklist_remove", params)
	if err != nil {
		return err
	}
	return nil
}

// BlacklistSet 设置黑名单
func (c *imClient) BlacklistSet(channelId string, channelType int, uids []string) error {
	params := map[string]interface{}{
		"channel_id":   channelId,
		"channel_type": channelType,
		"uids":         uids,
	}
	_, err := c.post("/channel/blacklist_set", params)
	if err != nil {
		return err
	}
	return nil
}

// WhitelistAdd 添加白名单
func (c *imClient) WhitelistAdd(channelId string, channelType int, uids []string) error {
	params := map[string]interface{}{
		"channel_id":   channelId,
		"channel_type": channelType,
		"uids":         uids,
	}
	_, err := c.post("/channel/whitelist_add", params)
	if err != nil {
		return err
	}
	return nil
}

// WhitelistRemove 移除白名单
func (c *imClient) WhitelistRemove(channelId string, channelType int, uids []string) error {
	params := map[string]interface{}{
		"channel_id":   channelId,
		"channel_type": channelType,
		"uids":         uids,
	}
	_, err := c.post("/channel/whitelist_remove", params)
	if err != nil {
		return err
	}
	return nil
}

// WhitelistSet 设置白名单
func (c *imClient) WhitelistSet(channelId string, channelType int, uids []string) error {
	params := map[string]interface{}{
		"channel_id":   channelId,
		"channel_type": channelType,
		"uids":         uids,
	}
	_, err := c.post("/channel/whitelist_set", params)
	if err != nil {
		return err
	}
	return nil
}

// MessageSend 发送消息
func (c *imClient) MessageSend(msg *MsgSendReq) (*MsgSendResp, error) {
	rsp, err := c.post("/message/send", msg)
	if err != nil {
		return nil, err
	}
	dataResult := gjson.Get(string(rsp), "data")
	messageID := dataResult.Get("message_id").Int()
	messageSeq := dataResult.Get("message_seq").Int()
	clientMsgNo := dataResult.Get("client_msg_no").String()
	return &MsgSendResp{
		MessageID:   messageID,
		MessageSeq:  uint32(messageSeq),
		ClientMsgNo: clientMsgNo,
	}, nil
}

// ChannelMessageAsync 频道消息同步
func (c *imClient) ChannelMessageAsync(uid, channelId string, channelType, start, end, limit, pullMode int) (*SyncChannelMessageResp, error) {
	params := map[string]interface{}{
		"login_uid":         uid,
		"channel_id":        channelId,
		"channel_type":      channelType,
		"start_message_seq": start,
		"end_message_seq":   end,
		"limit":             limit,
		"pull_mode":         pullMode,
	}
	resp, err := c.post("/channel/messagesync", params)
	if err != nil {
		return nil, err
	}
	var rsp SyncChannelMessageResp
	if err = json.Unmarshal(resp, &rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}

func (c *imClient) QuitUserDevice(uid string, deviceFlag int) error {
	params := map[string]interface{}{
		"uid":         uid,
		"device_flag": deviceFlag,
	}
	_, err := c.post("/user/device_quit", params)
	if err != nil {
		return err
	}
	return nil
}

func (c *imClient) GetChannelMaxSeq(channelId string, channelType int) (uint32, error) {
	params := map[string]interface{}{
		"channel_id":   channelId,
		"channel_type": channelType,
	}
	response, err := c.post("/channel/max_message_seq", params)
	if err != nil {
		return 0, err
	}
	var result struct {
		MessageSeq uint32 `json:"message_seq"`
	}
	if err = json.Unmarshal(response, &result); err != nil {
		return 0, err
	}
	return result.MessageSeq, nil
}

// SyncUserConversation 同步用户会话
func (c *imClient) SyncUserConversation(uid string, version int64, msgCount int64, lastMsgSeqs string, larges []*Channel) ([]*Conversation, error) {
	params := map[string]interface{}{
		"uid":           uid,
		"version":       version,
		"last_msg_seqs": lastMsgSeqs,
		"msg_count":     msgCount,
		"larges":        larges,
	}
	response, err := c.post("/conversation/sync", params)
	if err != nil {
		return nil, err
	}
	var conversations []*Conversation
	if err = json.Unmarshal(response, &conversations); err != nil {
		return nil, err
	}
	return conversations, nil
}

// ClearConversationUnread 清理未读消息
func (c *imClient) ClearConversationUnread(req ClearConversationUnreadReq) error {
	_, err := c.post("/conversations/setUnread", req)
	if err != nil {
		return err
	}
	return nil
}

// DeleteConversation 删除会话
func (c *imClient) DeleteConversation(req DeleteConversationReq) error {
	_, err := c.post("/conversations/delete", req)
	if err != nil {
		return err
	}
	return nil
}

// SyncMessage 消息同步
func (c *imClient) SyncMessage(uid string, messageSeq uint32, limit int) ([]*Message, error) {
	params := map[string]interface{}{
		"uid":         uid,
		"message_seq": messageSeq,
		"limit":       limit,
	}
	response, err := c.post("/message/sync", params)
	if err != nil {
		return nil, err
	}
	var messages []*Message
	if err = json.Unmarshal(response, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

// SyncMessageAck 同步IM消息回执
func (c *imClient) SyncMessageAck(uid string, lastMessageSeq uint32) error {
	params := map[string]interface{}{
		"uid":              uid,
		"last_message_seq": lastMessageSeq,
	}
	_, err := c.post("/message/syncack", params)
	if err != nil {
		return err
	}
	return nil
}

// GetWithChannelAndSeqs 同步IM消息回执
func (c *imClient) GetWithChannelAndSeqs(channelID string, channelType uint8, loginUID string, seqs []uint32) (*SyncChannelMessageResp, error) {
	params := map[string]interface{}{
		"channel_id":   channelID,
		"channel_type": channelType,
		"message_seqs": seqs,
		"login_uid":    loginUID,
	}
	resp, err := c.post("/messages", params)
	if err != nil {
		return nil, err
	}
	var rsp SyncChannelMessageResp
	if err = json.Unmarshal(resp, &rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// MessageSearch 消息搜索
func (c *imClient) MessageSearch(req MessageSearchReq) ([]*Message, error) {
	response, err := c.post("/message/search", req)
	if err != nil {
		return nil, err
	}
	var messages []*Message
	if err = json.Unmarshal(response, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}
