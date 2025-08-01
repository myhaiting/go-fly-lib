package im

// ---------- req  ----------

type PullMode int

const (
	PullModeDown PullMode = iota
	PullModeUp
)

// MessageHeader Message header
type MessageHeader struct {
	NoPersist int `json:"no_persist"` // Is it not persistent
	RedDot    int `json:"red_dot"`    // Whether to show red dot
	SyncOnce  int `json:"sync_once"`  // This message is only synchronized or consumed once
}

// Message 通知消息
type Message struct {
	Header       MessageHeader `json:"header"`              // 消息头
	Setting      uint8         `json:"setting"`             // 设置
	MessageID    int64         `json:"message_id"`          // 服务端的消息ID(全局唯一)
	MessageIDStr string        `json:"message_idstr"`       // 字符串类型服务端的消息ID(全局唯一)
	ClientMsgNo  string        `json:"client_msg_no"`       // 客户端消息唯一编号
	MessageSeq   uint32        `json:"message_seq"`         // 消息序列号 （用户唯一，有序递增）
	FromUID      string        `json:"from_uid"`            // 发送者UID
	ToUID        string        `json:"to_uid"`              // 接受者uid
	ChannelID    string        `json:"channel_id"`          // 频道ID
	ChannelType  uint8         `json:"channel_type"`        // 频道类型
	Timestamp    int32         `json:"timestamp"`           // 服务器消息时间戳(10位，到秒)
	Payload      []byte        `json:"payload"`             // base64消息内容
	StreamNo     string        `json:"stream_no,omitempty"` // 流编号
	Streams      []*StreamItem `json:"streams,omitempty"`   // 消息流
	Expire       uint32        `json:"expire"`              // 消息过期时间
	IsDeleted    int           `json:"is_deleted"`          // 是否已删除
	VoiceStatus  int           `json:"voice_status"`        // 语音状态 0.未读 1.已读
}

type StreamItem struct {
	StreamSeq   uint32 `json:"stream_seq"`    // 流序号
	ClientMsgNo string `json:"client_msg_no"` // 客户端消息唯一编号
	Blob        []byte `json:"blob"`          // 消息内容
}

// OfflineMessage 离线消息
type OfflineMessage struct {
	Message
	ToUIDs []string `json:"to_uids"` // 接收用户列表
}

// MsgSendReq 发送消息请求
type MsgSendReq struct {
	Header      MessageHeader `json:"header"`       // 消息头
	Setting     uint8         `json:"setting"`      // setting
	FromUID     string        `json:"from_uid"`     // 模拟发送者的UID
	ChannelID   string        `json:"channel_id"`   // 频道ID
	ChannelType uint8         `json:"channel_type"` // 频道类型
	StreamNo    string        `json:"stream_no"`    // 消息流号
	Subscribers []string      `json:"subscribers"`  // 订阅者 如果此字段有值，表示消息只发给指定的订阅者
	Payload     []byte        `json:"payload"`      // 消息内容
}

type MsgSendResp struct {
	MessageID   int64  `json:"message_id"`    // 消息ID
	ClientMsgNo string `json:"client_msg_no"` // 客户端消息唯一编号
	MessageSeq  uint32 `json:"message_seq"`   // 消息序号
}

type Channel struct {
	ChannelID   string `json:"channel_id"`   // 频道ID
	ChannelType uint8  `json:"channel_type"` // 频道类型
}

// Conversation 最近会话离线返回
type Conversation struct {
	ChannelID       string     `json:"channel_id"`         // 频道ID
	ChannelType     uint8      `json:"channel_type"`       // 频道类型
	Unread          int        `json:"unread"`             // 未读消息
	Timestamp       int64      `json:"timestamp"`          // 最后一次会话时间
	LastMsgSeq      int64      `json:"last_msg_seq"`       // 最后一条消息seq
	LastClientMsgNo string     `json:"last_client_msg_no"` // 最后一条客户端消息编号
	OffsetMsgSeq    int64      `json:"offset_msg_seq"`     // 偏移位的消息seq
	Version         int64      `json:"version"`            // 数据版本
	Recents         []*Message `json:"recents"`            // 最近N条消息
}

// ClearConversationUnreadReq 清除用户某个频道未读数请求
type ClearConversationUnreadReq struct {
	UID         string `json:"uid"`
	ChannelID   string `json:"channel_id"`
	ChannelType uint8  `json:"channel_type"`
	Unread      int    `json:"unread"`
	MessageSeq  uint32 `json:"message_seq"`
}

// DeleteConversationReq  DeleteConversationReq
type DeleteConversationReq struct {
	UID         string `json:"uid"`
	ChannelID   string `json:"channel_id"`   // 频道ID
	ChannelType uint8  `json:"channel_type"` // 频道类型
}

// SyncChannelMessageResp 同步频道消息返回
type SyncChannelMessageResp struct {
	StartMessageSeq uint32     `json:"start_message_seq"` // 开始序列号
	EndMessageSeq   uint32     `json:"end_message_seq"`   // 结束序列号
	PullMode        PullMode   `json:"pull_mode"`         // 拉取模式
	Messages        []*Message `json:"messages"`          // 消息数据
}

// MessageSearchReq 消息搜索请求
type MessageSearchReq struct {
	UID         string `json:"uid"` // 搜索的消息限定这某个用户内
	ChannelID   string `json:"channel_id"`
	ChannelType uint8  `json:"channel_type"`
	ContentType int    `json:"content_type"` // 正文类型
	Keyword     string `json:"keyword"`
}

// OnlineStatus 在线状态返回
type OnlineStatus struct {
	UID         string `json:"uid"`          // 在线用户uid
	DeviceFlag  uint8  `json:"device_flag"`  // 设备标记 0. APP 1.web
	LastOffline int    `json:"last_offline"` // 最后一次离线时间
	Online      int    `json:"online"`       // 是否在线
}
