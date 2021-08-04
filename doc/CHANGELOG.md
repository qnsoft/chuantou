# 版本变动日志

## Version 1.0.0

- 实现内网穿透
- 支持多端口穿透
- 服务端生成代理端口时回写给客户端，解决随机端口的问题

## Version 1.0.1
- 增加代理接入身份校验，解决服务端能被任何人使用的问题
- 支持客户端指定代理端口模式

## Version 1.0.2
- 使用延迟创建内网服务连接的方式，解决不停连接内网服务的问题
- 解决内网服务断开，访问服务仍未关闭的问题
- 去除随机端口访问模式
- 简化通讯协议

## Version 1.0.3
- 单服务支持多个连接，服务更稳定
- 通讯协议加入原端口

## Version 1.0.4
- 解决隧道桥中断，连接没有及时处理的问题
- 解决断线不能真正重连的问题

## Version 1.0.5
- 引入基于有效期的身份失效机制
- 增加版本校验机制，避免客户端与服务端版本不一致的问题

## Version 1.1.0
- 优化部分代码，使变得更可读
- 调整客户端端口映射配置，使变得更容易理解
- 修复客户端占用大量 CPU 的BUG（严重）
- 增加访问端口范围限制

## Version 1.2.0
- 修复由于客户端断开，未关闭监听造成启动后刷出大堆访问日志的BUG
- 重构服务端代码，调整服务端架构，解决服务端架构带来的`accept`问题
- 简化通讯协议，去除原端口参数
- 去除客户端参数“max-redial-times”，默认掉线无限重连
- 增加客户端参数“tunnel-count”，隧道条数，默认为1，范围\[1-5\]

## Version 1.2.1
- 优化代码，提升通讯效率

## Version 1.2.2
- 使用 go.mod 管理依赖
- 修复网络波动等因素会导致服务端不能正常提供服务的问题

## Version 1.3.0
- 通讯协议增加客户端ID标识，使用机器码
- 服务端增加访问端口校验，一个访问端口不能被重复占用
- 增加超时机制，每30分钟检测一遍隧道活性，失活则重建隧道

## Version 1.4.0
- 重构隧道活性检测机制，每1分钟检测隧道活性，容错率提高
- 修复端口占用不能准确报错的BUG


## TODO

- 服务端增加“最大端口数”配置，避免无限制开放端口
- 增加心跳机制检测服务是否通畅
- 通讯协议加密
- 增加命令行，查看端口列表
- 增加黑名单，支持屏蔽IP

通讯协议

- 通讯结果    1个字节(0: 成功，其他：失败)
- 版本号      4个字节
- 访问端口    4个字节
- 客户端ID    32位uuid
- Key        建议长度 6-16 字符串

协议最大长度不能超过 255

举例
1|1|13306|machineID|winshu

## 简易编译打包脚本

```shell script
# linux
git pull
go env -w GOPROXY=https://goproxy.cn,direct
go build -o netbus main.go
mkdir netbus_linux_amd64
mv netbus netbus_linux_amd64
cp config_demo.ini netbus_linux_amd64/config.ini
tar -zvcf netbus_linux_amd64.tar.gz netbus_linux_amd64
rm -rf netbus_linux_amd64
mv netbus_linux_amd64.tar.gz /mnt/d/

```

```bash
# windows
git pull
go build -o netbus.exe main.go
mkdir netbus_windows_amd64
move /Y netbus.exe netbus_windows_amd64
copy /Y config_demo.ini netbus_windows_amd64\config.ini
xcopy /Y doc\*.bat netbus_windows_amd64
if exist netbus_windows_amd64.zip ( del /Q netbus_windows_amd64.zip )
"C:/Program Files (x86)/WinRAR/WinRAR.exe" a netbus_windows_amd64.zip netbus_windows_amd64

```


