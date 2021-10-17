# 容器网络

### Linux 虚拟网络设备

我们都知道， Linux 实际是通过网络设备去操作和使用网卡的，系统装了 个网卡之后会为其生成一个网络设备实例，比如 eth0 。而随着网络虚拟化技术的发展， Linux 支持创建出虚拟化的设备，可以通过虚拟化设备的组合实现多种多样的功能和网络拓扑。常见的虚拟化设备 Veth Bridge 802 VLAN device TAP 这里主要介绍构建容器网络要用到的 Veth Bridge

#### Linuc Veth
Veth 是成对出现的虚拟网络设备，发送 Veth 一端虚拟设备的请求会从另一端的虚拟设备中发出。在容器的虚拟 场景中，经常会使用 Veth 连接不同的网络 Namespace 下。

``` bash
sudo ip netns add ns1

sudo ip netns add ns2

# 创建一对Veth
sudo ip link add veth0 type veth peer name veth1

# 分别将两个Veth移动到两个NameSpace中
sudo ip link set veth0 netns ns1

sudo ip link set veth2 netns ns2

# 去ns1的namspace中查看网络设备
sudo ip netns exec ns1 ip link

1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN mode DEFAULT group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
18: veth0@if17: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN mode DEFAULT group default qlen 1000
    link/ether ba:4f:21:0c:a1:1f brd ff:ff:ff:ff:ff:ff link-netnsid 0

```

ns1 和 ns2 的Namespace中，除了 loopback 设备以外就只看到了一个网络设备。当请求发送到这个虚拟网络设备 ，都会原封不动地从另一个网络 Namespace 的网络接口中出来。例如，给两端分别配置不同的地址后 ，向虚拟网络设备的一端发送请求，就能到达这个
虚拟网络设备对应的另一端。

``` bash
# 配置每个 veth 的网络地址和NameSpace的路由
sudo ip netns exec ns1 ifconfig veht0 172.18.0.2/24 up
sudo ip netns exec ns2 ifconfig veht1 172.18.0.3/24 up
sudo ip netns exec ns1 route add default dev veth0
sudo ip netns exec ns2 route add default dev veth1
# 通过 veth 一端出去的报 另一端能够直接收到
sudo ip netns exec ns1 ping -c 1 172.18.0.3

```

#### Linux bridge

Bridge 虚拟设备是用来桥接的的网络设备，它相当于现实世界中的交换机 可以连接不同的网络设备，当请求到达 Bridge 设备时，可以通过报文中的 Mac 地址进行广播或转发。例如，创建 Bridge 设备，来连接 Namespace 中的网络设备和宿主机上的网络

``` bash
# 创建Veth设备并将一端移入NAmespace
sudo ip netns add ns1
sudo ip link add veth0 type veth peer name veth1
sudo ip link set veth1 netns ns1
# 创建网桥
sudo vroctl addif br0 eth0
# 挂载网络设备
sudo brctl addif br0 eth0 # 这个得看自己的机器 想接哪个网卡了
sudo brctl addif br0 veth0
```

### Linuc路由表
路由表是 Linux 内核的一个模块，通过定义路由表来决定在某个网络 NameSpace 中包的流向，从而定义请求会到哪个网络设备上。

``` bash
sudo ip link set veth0 up
sudo ip link set br0 up
sudo ip netns exec ns1 ifconfig veth1 172.18.0.2/24 up
# 分别设置ns1网络空间的路由和宿主机上的路由
# default代表 0.0.0.0/0 即在 Net Namespace 中所有流量都经过 veth1 的网络设备流出
sudo ip netns exec ns1 route add default dev veth1
# 在宿主机上将172.18.0.0/24的网段请求路由到br0的网桥
sudo route add -net 172.18.0.0/24 dev br0
```

``` bash
eintr@localhost ~> ifconfig ens1
ens1: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 192.168.122.182  netmask 255.255.255.0  broadcast 192.168.122.255
        inet6 fe80::58c3:3365:c635:b02c  prefixlen 64  scopeid 0x20<link>
        ether 52:54:00:ce:42:d0  txqueuelen 1000  (Ethernet)
        RX packets 1356  bytes 849283 (829.3 KiB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 736  bytes 54814 (53.5 KiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

eintr@localhost ~> sudo ip netns exec ns1 ping -c 2 192.168.122.182 
PING 192.168.122.182 (192.168.122.182) 56(84) bytes of data.
64 bytes from 192.168.122.182: icmp_seq=1 ttl=64 time=0.236 ms
64 bytes from 192.168.122.182: icmp_seq=2 ttl=64 time=0.052 ms

--- 192.168.122.182 ping statistics ---
2 packets transmitted, 2 received, 0% packet loss, time 999ms
rtt min/avg/max/mdev = 0.052/0.144/0.236/0.092 ms

eintr@localhost ~> ping -c 2 172.18.0.2
PING 172.18.0.2 (172.18.0.2) 56(84) bytes of data.
64 bytes from 172.18.0.2: icmp_seq=1 ttl=64 time=0.198 ms
64 bytes from 172.18.0.2: icmp_seq=2 ttl=64 time=0.059 ms

--- 172.18.0.2 ping statistics ---
2 packets transmitted, 2 received, 0% packet loss, time 999ms
rtt min/avg/max/mdev = 0.059/0.128/0.198/0.070 ms
```

### Linuc iptables
iptables 是对 Linux 内核的 netfilter 模块进行操作和展示的工具，用来管理包的流动和转送。 iptables 定义了一套链式处理的结构，在网络包传输的各个阶段可以使用不同的策略对包进行加工、传送或丢弃。在容器虚拟化的技术中，经常会用到两种策略 MASQUERADE 和 DNAT ，用于容器和宿主机外部的网络通信。

#### MASQUERADE
itables 中的MASQUERADE 策略可以将请求包中的源地址转换策划那个一个网络设备的地址，比如之前介绍的那个 Namespace中的网络设备的地址是 172.18.0.2,这个地址虽然在宿主机上可以路由到br0的网桥，但是到达宿主机外部之后，是不知道如何路由到这个IP地址的，所以如果请求外部地址的花，需要先通过MASQUERADE策略将这个IP转换成宿主机出口网卡的IP

``` bash
# 打开IP转发
sudo sysctl -w net.ipv4.conf.all.forwarding=1
cat /proc/sys/net/ipv4/ip_forward

# 对 Namespace中发出的报添加网络地址转换
sudo iptables -t nat -A POSTROUTING -s 172.18.0.0/24 -o eth0 -j MASQUERADE
```

#### DNAT
iptables 中的 DNAT 策略也是做网络地址的转换，不过它是要更换目标地址，经常用于将内部网络地址的端口映射到外部去。比如，上面那个例子中的 Namespace 如果需要提供服务给宿主机之外的应用去请求要怎么办呢？外部应用没办法直接路由到 172.18.0.2 这个地址，这时候就可以用到 DNAT 策略。

``` bash
# 将到宿主机上80端口的请求转发到 Namespace的IP上
sudo iptables -t nat -A PREROUTING -p tcp -m tcp --dport 80 -j DNAT --to-desination 172.18.0.2:80
```

这样就可以把宿主机上 80 端口的 TCP 请求转发到 Namespace 中的地址 172.18.0.2:80 ，从而实现外部的应用调用。
