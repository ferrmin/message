package main

import (
	"bytes"
	"compress/zlib"
)

// encode 编码器(text/binary)
// binary一般可选以下两种方案
// ⑴ protobuf编码，html前端需要引入protobuf.js，并且设置WebSocket对象属性binaryType = 'arraybuffer'
// ⑵ binary+gzip/zlib编码，html前端需要引入pako.js，(可选)设置WebSocket对象属性binaryType = 'blob'
func encode(codec int, data []byte) []byte {
	switch codec {
	case 1:
		return data
	case 2:
		buff := new(bytes.Buffer)
		zw := zlib.NewWriter(buff)
		zw.Write(data)
		zw.Close()
		return buff.Bytes()
	}

	return data
}
