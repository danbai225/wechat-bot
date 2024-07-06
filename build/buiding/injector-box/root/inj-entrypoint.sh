#!/usr/bin/env bash
TARGET_AUTO_RESTART=${TARGET_AUTO_RESTART:-no}
TARGET_LOG_FILE=${TARGET_LOG_FILE:-/dev/null}
function run-target() {
    while :
    do
        $TARGET_CMD >${TARGET_LOG_FILE} 2>&1
        case ${TARGET_AUTO_RESTART} in
        false|no|n|0)
            exit 0
            ;;
        esac
    done
}
function run-qrServer() {
    nohup /home/app/qr-server >/dev/null 2>&1 &
}
/entrypoint.sh &
sleep 5
inject-monitor &
run-target &
run-qrServer &
wait
