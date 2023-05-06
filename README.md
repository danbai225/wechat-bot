# wechat-bot

运行在docker中的微信机器人支持批量部署代理设置。
文件发送与接收图片解析多语言sdk，http接口方便对接。

# docker-compose
```agsl
version: "3"
services:
  wechat-bot:
    image: danbai225/wechat-bot:latest
    container_name: wechat-bot
    restart: always
    ports:
      - "8080:8080"
      - "5555:5555"
      - "5556:5556"
      - "5900:5900"
    extra_hosts:
      - "dldir1.qq.com:127.0.0.1"
    volumes:
      - "./data:/home/app/data"
      - "./wxFiles:/home/app/WeChat Files"
    environment:
      #- PROXY_IP=1.1.1.115 #如果设置则使用代理
      - PROXY_PORT=7777
      - PROXY_USER=user
      - PROXY_PASS=pass
```

# use

```go
package main

import (
	logs "github.com/danbai225/go-logs"
	wechatbot "github.com/danbai225/wechat-bot"
)

func main() {
	client, err := wechatbot.NewClient("ws://serverIP:5555", "http://serverIP:5556")
	if err != nil {
		logs.Err(err)
		return
	}
	client.SetOnWXmsg(func(msg []byte, Type int, reply *wechatbot.Reply) {
		if Type == 1 {
			logs.Info(string(msg))
		}
	})
	select {}
}

```