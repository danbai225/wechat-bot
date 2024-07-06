FROM chisbread/wechat-service:latest
COPY target/root/ /
RUN sudo rm -r /payloads
COPY root/ /
RUN sudo chmod +x /*.sh && sudo chmod +x /bin/* && sudo chmod +x /payloads/*
ADD qr-server /home/app/
RUN sudo chmod +x /home/app/qr-server
ADD proxy.sh /home/app/
RUN sudo chmod +x /home/app/proxy.sh
RUN sudo chown -R app:app /drive_c && cp -r /drive_c/* /home/app/.wine/drive_c/
RUN sudo rm -rf /WeChatSetup*
RUN mkdir /home/app/data
RUN sudo ln -s /home/app/data /home/app/.wine/dosdevices/c:/data
ADD v.py /home/app/
RUN wine pip.exe install pymem
ENV HOOK_PROC_NAME=WeChat
ENV TARGET_AUTO_RESTART=yes
ENV TARGET_CMD=wechat-start
ENV INJ_CONDITION='[ "`ps -aux | grep funtool | grep -v grep`" != "" ] && exit 0'
ENTRYPOINT ["/inj-entrypoint.sh"]