from pymem import Pymem


def fix_version(pm: Pymem):
    WeChatWindll_base = 0
    for m in list(pm.list_modules()):
        path = m.filename
        if path.endswith("WeChatWin.dll"):
            WeChatWindll_base = m.lpBaseOfDll
            break

    # 这些是找到CE的标绿的内存地址偏移量
    ADDRS = [0x22300E0, 0x223D90C, 0x223D9E8, 0x2253E4C]

    for offset in ADDRS:
        addr = WeChatWindll_base + offset
        v = pm.read_uint(addr)
        print(v)
        if v == 0x63090a1b:  # 是3.9.10.27，已经修复过了
            continue
        elif v != 0x63060012:  # 不是 3.6.0.18 修复也没用，代码是hardcode的，只适配这一个版本
            raise Exception("别修了，版本不对，修了也没啥用。")

        pm.write_uint(addr, 0x63090a1b) # 改成要伪装的版本3.9.10.27，转换逻辑看链接

    print("好了，可以扫码登录了")


if __name__ == "__main__":
    try:
        pm = Pymem("WeChat.exe")
        fix_version(pm)
    except Exception as e:
        print(f"{e}，请确认微信程序已经打开！")

#pip install pymem 先安装pymem， 打开微信3.6.0.18， 然后运行 pip install pymem，即可登录