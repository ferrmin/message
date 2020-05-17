package main

import (
	"sync"

	"github.com/ferrmin/utils/redis"
	"github.com/go-ini/ini"
	"github.com/gorilla/websocket"
)

var (
	clients   sync.Map
	broadcast chan *broadcastEntity
	reply     chan *replyEntity
	redisCli  *redis.ConnPool
	nsqConf   NsqConf
	wsConf    WebSocketConf
)

type broadcastEntity struct {
	channel string
	message []byte
}

type replyEntity struct {
	c       *wsConn
	message []byte
}

type wsConn struct {
	conn  *websocket.Conn
	codec int //编码类型 (1:text; 2:binary|protobuf)
}

func init() {
	cfg, err := ini.Load("./conf/config.ini")
	if err != nil {
		panic(err)
	}

	nsqCfg := cfg.Section("nsq")
	nsqConf.LookupdAddr = nsqCfg.Key("lookupd_addr").Strings("|")
	nsqConf.MaxInFlight, _ = nsqCfg.Key("max_in_flight").Int()
	nsqConf.LookupdPollInterval, _ = nsqCfg.Key("lookupd_poll_interval").Int()
	nsqConf.ConsumerTopics = nsqCfg.Key("consumer_topics").Strings("|")
	nsqConf.Channel = nsqCfg.Key("channel").String()

	wsCfg := cfg.Section("websocket")
	wsConf.SrvAddr = wsCfg.Key("srv_addr").String()
	wsConf.ReadLimit, _ = wsCfg.Key("read_limit").Int64()
	wsConf.WriteTimeout, _ = wsCfg.Key("write_timeout").Int()

	redisCli = redis.InitRedisPool()

	broadcast = make(chan *broadcastEntity)
	reply = make(chan *replyEntity)
}

// NsqConf ...
type NsqConf struct {
	LookupdAddr         []string // 监听地址
	MaxInFlight         int      // 监听的最大节点数
	LookupdPollInterval int      // 重连间隔时间(秒)
	ConsumerTopics      []string // 消费者队列主题名称
	Channel             string   // 通道名称
}

// WebSocketConf ...
type WebSocketConf struct {
	SrvAddr      string // 服务地址
	ReadLimit    int64  // 接收数据长度限制(字节)
	WriteTimeout int    // 发送数据超时(秒)
}

// WsReq websocket请求
type WsReq struct {
	Req     string `json:"req"`     // 请求类型(ping|sub|unsub)
	Channel string `json:"channel"` // 数据频道
	Symbol  string `json:"symbol"`  // 交易对
	Period  string `json:"period"`  // 时间周期
	Begin   string `json:"begin"`   // 开始时间戳
	Offset  int    `json:"offset"`  // 偏移量
	Limit   int    `json:"limit"`   // 数据量
}

// Kline K线消息体
type Kline struct {
	Symbol string  `json:"symbol"` // 交易对
	Time   int64   `json:"time"`   // 时间戳
	Open   float64 `json:"open"`   // 开盘价
	Close  float64 `json:"close"`  // 关盘价
	High   float64 `json:"high"`   // 最高价
	Low    float64 `json:"low"`    // 最低价
	Num    float64 `json:"num"`    // 成交量
	Period string  `json:"period"` // 时间周期(1min 5min 15min 30min 1h 1d 1w 1m)
}

// Depth 盘口消息体
type Depth struct {
	Symbol string        `json:"symbol"` // 交易对
	Time   int64         `json:"time"`   // 时间戳
	Asks   []interface{} `json:"asks"`   // 卖
	Bids   []interface{} `json:"bids"`   // 买
}

// Asks 卖方消息体
type Asks struct {
	Price float64 `json:"price"` // 价格
	Num   float64 `json:"num"`   // 数量
}

// Bids 买方消息体
type Bids struct {
	Price float64 `json:"price"` // 价格
	Num   float64 `json:"num"`   // 数量
}

// Trades 实时成交消息体
type Trades struct {
	Symbol string  `json:"symbol"` // 交易对
	Type   int     `json:"type"`   // 1:buy|2:sell
	Time   int64   `json:"time"`   // 时间戳
	Price  float64 `json:"price"`  // 价格
	Num    float64 `json:"num"`    // 数量
}

// Latest 最新行情消息体
type Latest struct {
	Symbol string  `json:"symbol"` // 交易对
	Price  float64 `json:"price"`  // 最新价
	Rose   float64 `json:"rose"`   // 涨幅(%)
	High   float64 `json:"high"`   // 最高价
	Low    float64 `json:"low"`    // 最低价
	Num    float64 `json:"num"`    // 成交量
	Amount float64 `json:"amount"` // 成交额
	Time   int64   `json:"time"`   // 时间戳
}

// KlineScript k线脚本
const KlineScript = `
	local v = redis.call("ZCARD", KEYS[1])
	if v > 1050
	then
		redis.call("ZREMRANGEBYRANK", KEYS[1], 0, 50)
	end
	redis.call("ZREMRANGEBYSCORE", KEYS[1], ARGV[1], ARGV[1])
	redis.call("ZADD", KEYS[1], ARGV[1], ARGV[2])
	`

// TradesScript 实时成交脚本
const TradesScript = `
	local v = redis.call("ZCARD", KEYS[1])
	if v > 150
	then
		redis.call("ZREMRANGEBYRANK", KEYS[1], 0, 50)
	end
	redis.call("ZADD", KEYS[1], ARGV[1], ARGV[2])
	`
