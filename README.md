# SmartAgent Agent


SmartAgent 采用C/S架构的模型来运行，两者之间采用wss协议保障传输安全性，为保障高性能运行，协议中的数据流采用protobuf协议封装进行传输。为提高可扩展性，在Agent端使用多进程的方式运行多种插件，来提供业务方的扩展能力。

<br>

### 服务端
Server为SmartAgent的控制端，负责控制所有主机极其运行插件。

<br>

### 客户端
Agent为SmartAgent的受控端，目前通过插件的方式已支持：远程命令（脚本）执行、获取主机文件列表、文件上传下载、远程命令行等。Agent端支持对执行插件的CPU和内存限制，以此来保障Agent端不会因为自生原因而影响宿主机上的其他服务。

规划中插件能力：  
- Docker插件  
- 自动化机器人（RPA）插件  


<br>

## 产品特色
### 安全性
SmartAgent服务端在接受新的链接后会等待客户端的握手消息，其中包含客户端所在主机的操作系统、CPU、内存、主机名等基础信息。若服务端在5秒内无法收到客户端的握手消息或者握手消息格式有误时，服务端将会主动断开链接。
<br>

### 插件化
为丰富SmartAgent自身功能，SmartAgent提供了插件化的能力。在服务器端接收到客户端的握手消息并确认无误后，将会根据客户端的操作系统分发对应操作系统的可执行插件。客户端在启动时会监听一个gRPC端口（10010），该端口被监听在127.0.0.1这个IP地址上，该服务用于将SmartAgent服务端所发送的流量转发给插件进行处理。应此该客户机上的Agent只与当前主机上的插件进行通信。
<br>

### 跨平台
SmartAgent主程序采用go语言进行开发，应此兼容市面上大部分操作系统，如centos、redhat、debian、ubuntu、suse、solaris、xp、win2003、win2008、2012、2016、2019等，go语言本身要求linux内核版本号不低于2.6.23，理论上SmartAgent支持该内核版本以上的所有linux操作系统，包括嵌入式Arm、MIPS平台等。
<br>

### Restful接口
Server端提供了完整的API接口获取Agent列表、主机信息、任务执行等操作，可以将SmartAgent作为底层通信组件便于集成到上层分布式业务系统中。



<br>
<br>
<br>

SmartAgent Client   
https://github.com/jkstack/SmartAgent

<br>

SmartAgent Server   
https://github.com/jkstack/SmartAgent-server

<br>
<br>

SmartAgent 开源站点<br>

http://open.jkstack.com


SmartAgent 用户微信群

<img src="wechat_QR.jpg" height=200px weight=200px>