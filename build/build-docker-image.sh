#!/usr/bin/env bash
mkdir buiding || true
cp funtool/funtool_3.6.0.18-1.0.0015非注入版.exe buiding/injector-box/root/bin/
cp funtool/inject-dll buiding/injector-box/root/bin/
cp funtool/inject-monitor buiding/injector-box/root/bin/
cd buiding/injector-box
##api

sudo docker build -t danbai225/wechat-bot:latest .
#sudo docker push  danbai225/wechat-bot:latest

