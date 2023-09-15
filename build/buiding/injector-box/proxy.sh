#!/usr/bin/env bash
 #设置代理
function setProxy() {
  wxWid=$(xdotool search '微信')
  xdotool windowmove $wxWid 0 0
  sleep 0.5
  if [ -n "$PROXY_IP" ]; then
    xdotool mousemove 185 128
    sleep 0.5
    xdotool click 1
    sleep 0.5
    xdotool key Tab
    sleep 0.5
    xdotool key Tab
    sleep 0.5
    xdotool key Return
    xdotool mousemove 185 128
    sleep 0.5
    xdotool click 1
    sleep 0.5
    xdotool mousemove 222 237
    sleep 0.5
    xdotool click 1
    sleep 0.5
    xdotool mousemove 130 170
    sleep 0.5
    xdotool click 1
    sleep 0.5
    xdotool type $PROXY_IP
    sleep 0.5
    xdotool key Tab
    sleep 0.5
    xdotool type $PROXY_PORT
    sleep 0.5
    xdotool key Tab
    sleep 0.5
    xdotool type $PROXY_USER
    sleep 0.5
    xdotool key Tab
    sleep 0.5
    xdotool type $PROXY_PASS
    sleep 0.5
    xdotool mousemove 140 330
    sleep 0.5
    xdotool mousedown 1 && sleep 1 && xdotool mouseup 1
    touch ~/.proxy
    echo 'ok' >> ~/.proxy
  fi
}
if [ ! -e ~/.proxy ]; then
  echo "file not exist"
  setProxy
fi