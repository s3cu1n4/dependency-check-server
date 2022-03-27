# dependency-check-server
dependency-check-server 通过使用http服务，将jar包发送到服务端进行检测，如果你的dependency-check由于某些限制不能正常集成到Jenkins或者SonarQube中，可以考虑通过这种方式来实现jar包依赖组件的检测。

# Quick Start 

# server

## Docker

```bash
docker pull s3cu1n4/dcs:latest
   ```

```bash
docker run -p 5800:5800 -d s3cu1n4/dcs:latest
   ```

为方便检测结果检索，本服务支持将检测结果发送到阿里云日志服务，修改配置文件即可开启阿里云日志服务保存检测结果。

配置文件使用yaml语法，参数可参考：
```
DCServer:
  listenPort: "5800"
  aliyunlog: false

# aliyunlog 配置为 true 时，请配置Aliyun日志服务的各项配置


Aliyun:
  endpoint: ""
  ak: ""
  sk: ""
  Slsproject: ""
  Logstore: ""
```

可在宿主机修改配置文件，通过docker 启动参数将宿主机目录挂载到容器内即可
如配置文件在 /root/dependency-check-server/conf/ 目录下时，启动参数可改成如下：


```bash
docker run -p 5800:5800 -v /root/dependency-check-server/conf:/src/conf/ -d s3cu1n4/dcs:latest
   ```

## 查看检测结果
1、本服务默认在5800端口启动了http服务，可在浏览器内访问 `http://youserverip:5800/report/` 直接查看检测结果 `注意：该端口无认证措施，请注意安全防护，避免开启互联网访问`；

2、如果开启了阿里云日志服务，可通过阿里云日志服务直接查看检测结果。



***

# client 

client 通过fsnotify监控目录下的文件生成情况，当检测到jar文件写入时，会将jar通过http服务发送到服务端进行检测

client 端通过 -c 参数加载配置文件，配置文件参数如下：

```
client:
  serveraddr: "http://youserverip:5800/uploadjar"
  monitorjarpath: "./"
  jartemppath: "/tmp/jar/"
```
client 启动命令
```bash
./linux_server -c conf/conf.yaml
   ```

可使用nohup在后台运行，或者自行配置开机启动

```bash
nohup ./linux_server -c conf/conf.yaml &
   ```






## Contribution

欢迎提交 PR、Issues。