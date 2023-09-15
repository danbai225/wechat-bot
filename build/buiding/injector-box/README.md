# Injector Box
- Happy Hacking!
- 其实这是一个拼图游戏, 如果你能还原出原本的功能. 请将过程发送到我的邮箱.
- 注意:
  - 使用方式: 本项目旨在研究如何在wine环境下进行*win32远线程注入*, 不承诺具体的应用方式
  - 特殊假设: 需要与其它应用程序payloads结合使用
  - 版本依赖: 一般情况下, dll需要搭配指定版本的应用程序
## Riddle
```
git clone https://github.com/\* target

wget -P target/root https://<Any>/\*.exe
wget https://<Any>/\*.dll -o auto.dll

sudo docker build -t chisbread/injector-box:latest .

sudo docker run -it --name injector-box \
    -e HOOK_PROC_NAME=\* \
    -e HOOK_DLL=auto.dll \
    -e TARGET_CMD=\*-start \
    -v $PWD/auto.dll:/home/app/.wine/drive_c/injector/auto.dll \
    -p 8080:8080 -p 5555:5555 \
    chisbread/injector-box:latest
```


## Hints
- https://github.com/ChisBread/wechat-box
- https://github.com/cixingguangming55555/wechat-bot
- https://github.com/ChisBread/wine-box

## docker-compose
```yaml
version: "3.3"

services:
    injector-box:
        image: "chisbread/injector-box:latest"
        restart: unless-stopped
        container_name: "injector-box"
        environment:
            HOOK_PROC_NAME: 'WeProcName'
            HOOK_DLL: 'your.dll'
            TARGET_CMD: 'target-start'
            #optional INJMON_LOG_FILE: "/dev/null"
        ports:
            - "8080:8080" # noVNC
            - "5555:5555" # Hook API
        volumes:
            - "/path/to/your.dll:/home/app/.wine/drive_c/injector/your.dll"
            - "<path>:/home/app/.wine/drive_c/users/app/AppData/"
            - "<path>:/home/app/Some Files/" 
        tty: true

```