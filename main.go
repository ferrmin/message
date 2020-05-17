package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ferrmin/utils/helpers"
	"github.com/ferrmin/utils/logger"
	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 10 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	logger.Init()

	go consumer()
	go queueSender()

	http.HandleFunc("/message", handler)
	http.HandleFunc("/message/debug", debug)
	log.Fatal(http.ListenAndServe(wsConf.SrvAddr, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	//binary格式
	upgrade(w, r, 2)
}

func debug(w http.ResponseWriter, r *http.Request) {
	//text格式
	upgrade(w, r, 1)
}

func upgrade(w http.ResponseWriter, r *http.Request, codec int) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("websocket.Upgrade error: %v", err)
		return
	}

	c := &wsConn{
		codec: codec,
		conn:  conn,
	}

	defer func() {
		clients.Delete(c)
		conn.Close()
	}()

	// 接收数据长度限制
	conn.SetReadLimit(wsConf.ReadLimit)

	for {
		req := new(WsReq)
		err := conn.ReadJSON(&req)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure) {
				logger.Errorf("|%s| ReadJSON error: %v", helpers.GetRemoteIP(r), err)
			}
			return
		}

		//记录请求(测试)
		logger.Infof("|%s| {%s} {%s} {%s} {%s}", helpers.GetRemoteIP(r), req.Req, req.Channel, req.Symbol, req.Period)

		switch req.Req {
		case "ping":
			// 处理ping
			write(c, `{"channel":"pong","data":"`+req.Period+`"}`)
		case "unsub":
			// 处理退订
			unsubscribe(c, req.Channel)
		case "sub":
			// 处理订阅
			subscribe(c, req.Channel, req.Symbol, req.Period)
		default:
			// 拉模式
			switch req.Channel {
			case "market.kline":
				if req.Req == "0" {
					// 首次请求
					loadKline(c, req.Symbol, req.Period, 200)
				} else {
					// 最新一条
					loadKline(c, req.Symbol, req.Period, 1)
				}
			case "market.trades":
				loadTrades(c, req.Symbol, 100)
			case "market.depth":
				loadDepth(c, req.Symbol)
			case "market.latest":
				loadLatest(c)
			case "market.range.kline":
				// 范围拉取k线数据
				loadRangeKline(c, req.Offset, req.Limit, req.Symbol, req.Period, req.Begin)
			}
		}
	}
}

func (h *NsqHandler) push(msgBody []byte) {
	broadcast <- &broadcastEntity{channel: h.topicName, message: msgBody}
}

func (h *NsqHandler) store(msgBody []byte) {
	switch h.topicName {
	case "market.kline":
		mess := Kline{}
		if err := json.Unmarshal(msgBody, &mess); err != nil {
			logger.Errorf("redis market.kline message %s json.Unmarshal error: %v", string(msgBody), err)
		} else {
			key := h.topicName + ":" + mess.Symbol + ":" + mess.Period
			_, err := redisCli.LuaScriptZset(KlineScript, key, mess.Time, string(msgBody))
			if err != nil {
				logger.Errorf("market.kline LuaScriptZset error: %v", err)
			}
		}
	case "market.depth":
		mess := Depth{}
		if err := json.Unmarshal(msgBody, &mess); err != nil {
			logger.Errorf("redis market.depth message %s json.Unmarshal error: %v", string(msgBody), err)
		} else {
			redisCli.Set(h.topicName+":"+mess.Symbol, string(msgBody))
		}
	case "market.trades":
		mess := Trades{}
		if err := json.Unmarshal(msgBody, &mess); err != nil {
			logger.Errorf("redis market.trades message %s json.Unmarshal error: %v", string(msgBody), err)
		} else {
			_, err := redisCli.LuaScriptZset(TradesScript, h.topicName+":"+mess.Symbol, mess.Time, string(msgBody))
			if err != nil {
				logger.Errorf("market.trades LuaScriptZset error: %v", err)
			}
		}
	case "market.latest":
		mess := Latest{}
		if err := json.Unmarshal(msgBody, &mess); err != nil {
			logger.Errorf("redis market.latest message %s json.Unmarshal error: %v", string(msgBody), err)
		} else {
			redisCli.Hset(h.topicName, mess.Symbol, string(msgBody))
		}
	}
}

func unsubscribe(c *wsConn, channel string) {
	val, ok := clients.Load(c)
	if ok {
		array := val.([5]string)

		switch channel {
		case "market.kline":
			array[0] = ""
			array[1] = ""
		case "market.trades":
			array[2] = ""
		case "market.depth":
			array[3] = ""
		case "market.latest":
			array[4] = ""
		}

		if array == [5]string{} {
			clients.Delete(c)
		} else {
			clients.Store(c, array)
		}
	}

	write(c, `{"channel":"unsub","data":{},"status":"ok","error_msg":""}`)
}

func subscribe(c *wsConn, channel, symbol, period string) {
	//The structure of storage
	//key   value
	//{c}   {period,symbol,trades,depth,latest}

	var array [5]string
	val, ok := clients.Load(c)
	if ok {
		array = val.([5]string)
	}

	switch channel {
	case "market.kline":
		{
			loadKline(c, symbol, period, 200)
			array[0] = period
			array[1] = symbol
		}
	case "market.trades":
		{
			loadTrades(c, symbol, 100)
			array[2] = symbol
		}
	case "market.depth":
		{
			loadDepth(c, symbol)
			array[3] = symbol
		}
	case "market.latest":
		{
			loadLatest(c)
			array[4] = "latest"
		}
	}

	clients.Store(c, array)
}

func loadKline(c *wsConn, symbol, period string, n int) {
	klineData, _ := redisCli.Zrange("market.kline:"+symbol+":"+period, n*-1, -1)
	klineStr := "[" + strings.Join(klineData, ",") + "]"
	write(c, `{"channel":"market.history.kline","data":`+klineStr+`,"status":"ok","error_msg":""}`)
}

func loadTrades(c *wsConn, symbol string, n int) {
	tradesData, _ := redisCli.Zrange("market.trades:"+symbol, n*-1, -1)
	tradesStr := "[" + strings.Join(tradesData, ",") + "]"
	write(c, `{"channel":"market.history.trades","data":`+tradesStr+`,"status":"ok","error_msg":""}`)
}

func loadDepth(c *wsConn, symbol string) {
	depthData, _ := redisCli.Get("market.depth:" + symbol)
	if len(depthData) == 0 {
		depthData = "{}"
	}
	write(c, `{"channel":"market.history.depth","data":`+depthData+`,"status":"ok","error_msg":""}`)
}

func loadLatest(c *wsConn) {
	latestData, _ := redisCli.Hgetall("market.latest")
	arrays := make([]string, 0, len(latestData))
	for _, item := range latestData {
		arrays = append(arrays, item)
	}
	write(c, `{"channel":"market.history.latest","data":[`+strings.Join(arrays, ",")+`],"status":"ok","error_msg":""}`)
}

func loadRangeKline(c *wsConn, offset, limit int, symbol, period, begin string) {
	if limit <= 0 {
		limit = 1
	} else if limit > 200 {
		limit = 200
	}

	data, _ := redisCli.Zrevrangebyscore("market.kline:"+symbol+":"+period, "("+begin, "-inf", offset, limit)
	klineStr := "[" + strings.Join(data, ",") + "]"
	write(c, `{"channel":"market.range.kline","data":`+klineStr+`,"status":"ok","error_msg":""}`)
}

func write(c *wsConn, data string) {
	reply <- &replyEntity{c: c, message: []byte(data)}
}

func queueSender() {
	for {
		select {
		case entity := <-broadcast:
			clients.Range(func(cli, val interface{}) bool {
				var sendMess string
				msg := `{"channel":"` + entity.channel + `","data":` + string(entity.message) + `}`

				switch entity.channel {
				case "market.kline":
					mess := Kline{}
					if err := json.Unmarshal(entity.message, &mess); err == nil {
						if val.([5]string)[0] == mess.Period && val.([5]string)[1] == mess.Symbol {
							sendMess = msg
						}
					}
				case "market.trades":
					mess := Trades{}
					if err := json.Unmarshal(entity.message, &mess); err == nil {
						if val.([5]string)[2] == mess.Symbol {
							sendMess = msg
						}
					}
				case "market.depth":
					mess := Depth{}
					if err := json.Unmarshal(entity.message, &mess); err == nil {
						if val.([5]string)[3] == mess.Symbol {
							sendMess = msg
						}
					}
				case "market.latest":
					mess := Latest{}
					if err := json.Unmarshal(entity.message, &mess); err == nil {
						if val.([5]string)[4] == "latest" {
							sendMess = msg
						}
					}
				}

				if len(sendMess) > 0 {
					client := cli.(*wsConn)
					data := encode(client.codec, []byte(sendMess))
					client.conn.SetWriteDeadline(time.Now().Add(time.Duration(wsConf.WriteTimeout) * time.Second))
					err := client.conn.WriteMessage(client.codec, data)
					if err != nil {
						logger.Errorf("WriteMessage(broadcast) error: %v", err)
						client.conn.Close()
						clients.Delete(cli)
					}
				}

				return true
			})
		case entity := <-reply:
			data := encode(entity.c.codec, entity.message)
			entity.c.conn.SetWriteDeadline(time.Now().Add(time.Duration(wsConf.WriteTimeout) * time.Second))
			err := entity.c.conn.WriteMessage(entity.c.codec, data)
			if err != nil {
				logger.Errorf("WriteMessage(reply) error: %v", err)
				entity.c.conn.Close()
			}
		}
	}
}

// NsqHandler ...
type NsqHandler struct {
	topicName string
}

// HandleMessage ...
func (h *NsqHandler) HandleMessage(message *nsq.Message) error {
	go h.store(message.Body)
	h.push(message.Body)
	return nil
}

func consumer() {
	config := nsq.NewConfig()
	config.MaxInFlight = nsqConf.MaxInFlight                                              //连接的节点数量
	config.LookupdPollInterval = time.Duration(nsqConf.LookupdPollInterval) * time.Second //设置重连时间

	for _, topic := range nsqConf.ConsumerTopics {
		c, err := nsq.NewConsumer(topic, nsqConf.Channel, config)
		if err != nil {
			logger.Errorf("NewConsumer error: %v", err)
			return
		}

		c.SetLogger(nil, nsq.LogLevelDebug) //屏蔽系统日志
		c.AddHandler(&NsqHandler{topicName: topic})

		if err := c.ConnectToNSQLookupds(nsqConf.LookupdAddr); err != nil {
			logger.Errorf("ConnectToNSQLookupds error: %v", err)
		}
	}

	select {}
}
