# 消息推送系统
[![Build Status](https://travis-ci.org/ferrmin/message.svg?branch=master)](https://travis-ci.org/ferrmin/message)


## 推模式
> 一、订阅/取消订阅(sub/unsub) K线、实时成交、盘口消息(备注：symbol为交易对；period为时间周期如：1min 5min 15min 30min 1h 1d 1w 1m)

```
//订阅K线消息
{
    "req":"sub",
    "channel":"market.kline",
    "symbol":"BTC/USDT",
    "period":"1min"
}

//订阅实时成交消息
{
    "req":"sub",
    "channel":"market.trades",
    "symbol":"BTC/USDT"
}

//订阅盘口消息
{
    "req":"sub",
    "channel":"market.depth",
    "symbol":"BTC/USDT"
}
```

```
//响应订阅(成功)
{
    "channel":"market.history.kline",
    "data":[
        {"symbol":"BTC/USDT","time":1587744060,"open":7500.01,"close":7800.21,"high":8000.18,"low":7100.61,"num":0.0212,"period":"1min"},
        {"symbol":"BTC/USDT","time":1587744120,"open":7510.01,"close":7810.21,"high":8000.18,"low":7100.61,"num":0.1061,"period":"1min"},
        {"symbol":"BTC/USDT","time":1587744180,"open":7520.01,"close":7820.21,"high":8000.18,"low":7100.61,"num":0.2020,"period":"1min"}
    ],
    "status":"ok",
    "error_msg":""
}

//响应订阅(成功)
{
    "channel":"market.history.trades",
    "data":[
        {"symbol":"BTC/USDT","type":1,"price":7600.01,"num":0.0001,"time":1587744181000},
        {"symbol":"BTC/USDT","type":2,"price":7600.01,"num":0.0001,"time":1587744182000},
        {"symbol":"BTC/USDT","type":1,"price":7600.01,"num":0.0001,"time":1587744183000}
    ],
    "status":"ok",
    "error_msg":""
}

//响应订阅(成功)
{
    "channel":"market.history.depth",
    "data":{
        "time":1553651558,"symbol":"BTC/USDT",
        "asks":[[7611.01,0.0001],[7612.01,0.0001],[7613.01,0.0001]],
        "bids":[[7511.01,0.0001],[7512.01,0.0001],[7513.01,0.0001]]
    },
    "status":"ok",
    "error_msg":""
}
```

```
//订阅成功后会自动推送新数据(先后顺序不定) ⇩⇩⇩

//k线消息
{
    "channel":"market.kline",
    "data":{"time":1587744180,"symbol":"BTC/USDT","open":7510.01,"close":7810.21,"high":8000.18,"low":7100.61,"num":0.1061,"period":"1min"}
}

//实时成交消息
{
    "channel":"market.trades",
    "data":{"symbol":"BTC/USDT","type":1,"price":7511.01,"num":0.0001,"time":1587744186000}
}

//盘口消息
{
    "channel":"market.depth",
    "data":{
        "time":1587744180,"symbol":"BTC/USDT",
        "asks":[[7611.01,0.0001],[7612.01,0.0001],[7613.01,0.0001]],
        "bids":[[7511.01,0.0001],[7512.01,0.0001],[7513.01,0.0001]]
    }
}
```

```
//响应订阅(失败)
{   
    "channel":"market.history.kline",
    "data":[],
    "status":"error",
    "error_msg":"internal error"
}

{   
    "channel":"market.history.trades",
    "data":[],
    "status":"error",
    "error_msg":"internal error"
}

{   
    "channel":"market.history.depth",
    "data":{},
    "status":"error",
    "error_msg":"internal error"
}
```

> 二、订阅/取消订阅(sub/unsub) 最新行情消息

```
//发送
{
    "req":"sub",
    "channel":"market.latest"
}
```

```
//响应订阅(成功)
{   
    "channel":"market.history.latest",
    "data":[
        {"symbol":"BTC/USDT","price":7511.01,"rose":-0.0249,"high":8000.18,"low":7100.61,"num":16.0011,"amount":666666,"time":1587744180},
        {"symbol":"ETH/USDT","price":180.06,"rose":0.0146,"high":200.01,"low":166.07,"num":66.1011,"amount":888888,"time":1587744180}
    ],
    "status":"ok",
    "error_msg":""
}
```

```
//订阅成功后会自动推送新行情数据 ⇩⇩⇩

//最新行情消息
{
    "channel":"market.latest",
    "data":{"symbol":"BTC/USDT","price":7511.01,"rose":-0.0249,"high":8000.18,"low":7100.61,"num":16.0011,"amount":666666,"time":1587744180}
}
```

```
//响应订阅(失败)
{   
    "channel":"market.history.latest",
    "data":[],
    "status":"error",
    "error_msg":"internal error"
}
```

> 三、发送心跳(注意：超过一分钟没数据传输，请发送心跳包[每10秒发一次]，以免断开连接)

```
//发送
{
    "req":"ping",
    "period":"1587744180"
}
```

```
//返回
{   
    "channel":"pong",
    "data":"1587744180"
}
```

> 四、取消订阅(针对以上第一、第二点 req为unsub 时)统一返回

```
//返回
{  
    "channel":"unsub",
    "data":{},
    "status":"ok",
    "error_msg":"" 
}
```

## 拉模式
> 一、请求K线、盘口、实时成交数据(x秒请求一次，注意：req为"0"表示首次请求，留空为后续只请求最新数据)

```
//发送
{
    "req":"0",
    "channel":"market.kline",
    "symbol":"BTC/USDT", 
    "period":"1min"
}

//发送
{
    "channel":"market.trades",
    "symbol":"BTC/USDT"
}

//发送
{
    "channel":"market.depth",
    "symbol":"BTC/USDT"
}
```

```
//返回(成功)
{
    "channel":"market.history.kline",
    "data":[
        {"symbol":"BTC/USDT","time":1587744060,"open":7500.01,"close":7800.21,"high":8000.18,"low":7100.61,"num":0.0212,"period":"1min"},
        {"symbol":"BTC/USDT","time":1587744120,"open":7510.01,"close":7810.21,"high":8000.18,"low":7100.61,"num":0.1061,"period":"1min"},
        {"symbol":"BTC/USDT","time":1587744180,"open":7520.01,"close":7820.21,"high":8000.18,"low":7100.61,"num":0.2020,"period":"1min"}
    ],
    "status":"ok",
    "error_msg":""
}

{
    "channel":"market.history.trades",
    "data":[
        {"symbol":"BTC/USDT","type":1,"price":7600.01,"num":0.0001,"time":1587744181000},
        {"symbol":"BTC/USDT","type":2,"price":7600.01,"num":0.0001,"time":1587744182000},
        {"symbol":"BTC/USDT","type":1,"price":7600.01,"num":0.0001,"time":1587744183000}
    ],
    "status":"ok",
    "error_msg":""
}

{
    "channel":"market.history.depth",
    "data":{
        "time":1587744180,"symbol":"BTC/USDT",
        "asks":[[7611.01,0.0001],[7612.01,0.0001],[7613.01,0.0001]],
        "bids":[[7511.01,0.0001],[7512.01,0.0001],[7513.01,0.0001]]
    },
    "status":"ok",
    "error_msg":""
}
```

```
//返回(失败)
{   
    "channel":"market.history.kline",
    "data":[],
    "status":"error",
    "error_msg":"internal error"
}

{   
    "channel":"market.history.trades",
    "data":[],
    "status":"error",
    "error_msg":"internal error"
}

{   
    "channel":"market.history.depth",
    "data":{},
    "status":"error",
    "error_msg":"internal error"
}
```

> 二、请求最新行情数据(x秒请求一次)

```
//发送
{
    "channel":"market.latest"
}
```

```
//返回(成功)
{   
    "channel":"market.history.latest",
    "data":[
        {"symbol":"BTC/USDT","price":7511.01,"rose":-0.0249,"high":8000.18,"low":7100.61,"num":16.0011,"amount":666666,"time":1587744180},
        {"symbol":"ETH/USDT","price":180.06,"rose":0.0146,"high":200.01,"low":166.07,"num":66.1011,"amount":888888,"time":1587744180}
    ],
    "status":"ok",
    "error_msg":""
}
```

```
//返回(失败)
{   
    "channel":"market.latest",
    "data":[],
    "status":"error",
    "error_msg":"internal error"
}
```

## 分批拉取K线历史数据

支持图表缩放、拖动显示更多历史数据(max:1000)

```
//发送(备注：begin为开始时间戳；offset为偏移量(offset=limit*(page-1))；limit为拉取数据量(可以自定义数据量[1~200])
{
    "channel":"market.range.kline",
    "symbol":"BTC/USDT",
    "period":"1min",
    "begin":"1587744240",
    "offset":0,
    "limit":100
}
```

```
//返回(成功)
{   
    "channel":"market.range.kline",
    "data":[
        {"symbol":"BTC/USDT","time":1587744180,"open":7520.01,"close":7820.21,"high":8000.18,"low":7100.61,"num":0.2020,"period":"1min"},
        {"symbol":"BTC/USDT","time":1587744120,"open":7510.01,"close":7810.21,"high":8000.18,"low":7100.61,"num":0.1061,"period":"1min"},
        {"symbol":"BTC/USDT","time":1587744060,"open":7500.01,"close":7800.21,"high":8000.18,"low":7100.61,"num":0.0212,"period":"1min"}
    ],
    "status":"ok",
    "error_msg":""
}
```

```
//返回(失败)
{   
    "channel":"market.range.kline",
    "data":[],
    "status":"error",
    "error_msg":"internal error"
}
```

## JavaScript对接示例代码

```
//需要引入pako.js
//<script src="https://cdn.staticfile.org/pako/1.0.10/pako.min.js"></script>

var ws = new WebSocket('ws://192.168.3.7:8080/message');
//debug模式，支持浏览器直接查看文本数据
//var ws = new WebSocket('ws://192.168.3.7:8080/message/debug');

    ws.onopen = function() {
        //订阅k线消息
        ws.send('{"req":"sub","channel":"market.kline","symbol":"BTC/USDT","period":"1min"}');
        //订阅最新行情消息
        ws.send('{"req":"sub","channel":"market.latest"}');
    };
    ws.onmessage = function(evt) {
        processData(evt.data);
    };
    ws.onclose = function() {
        console.log('websocket断开');
    };
    ws.onerror = function() {
        console.log('websocket错误');
    };

function processData(data) {
    //心跳
    heartCheck.start();

    if (typeof data === "string") {  
        //文本格式
        console.log(JSON.parse(data));
    } else {
        //二进制带压缩格式
        var reader = new FileReader();
        reader.readAsArrayBuffer(data);

        reader.onload = function (e) {
            var blobBuff = new Uint8Array(e.target.result);
            var msg = pako.inflate(blobBuff, {to: 'string'});
            console.log(JSON.parse(msg));
        };
    }
}

var heartCheck = {
    timeout: 10000, //10秒发一次
    timer: null,
    start: function(){
        this.timer && clearTimeout(this.timer);
        this.timer = setTimeout(function(){
            ws.send('{"req":"ping","period":"'+(new Date()).valueOf()+'"}');
        }, this.timeout)
    }
};
```