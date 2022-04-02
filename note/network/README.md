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

sudo ip link set veth1 netns ns2

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
sudo ip netns exec ns1 ifconfig veth0 172.18.0.2/24 up
sudo ip netns exec ns2 ifconfig veth1 172.18.0.3/24 up
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
sudo brctl addbr br0
# 挂载网络设备
sudo brctl addif br0 eth0 # 这个得看自己的机器 想接哪个网卡了 不建议接主网络
sudo brctl addif br0 veth0
```

### Linux路由表
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

## 开始构建!

``` bash
Docker-ECE network create --subnet 192.168.0.0./24 --driver bridge testbridge
```

通过 Bridge 的网络驱动创建一个网络，网段是 192.168.0.0/24 ，网络驱动是 Bridge

![img](./6.5)

``` go
package network

import (
	"encoding/json"
	"net"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

var (
	defaultNetworkPath = "/var/run/docker-ece/network/"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*Network{}
)

type Network struct {
	Name    string     // 网络名
	IpRange *net.IPNet // 地址段
	Driver  string     // 网络驱动名
}

// 网络端点
// 网络端点是用于连接容器和网络的，保证容器内部与网络的通信
type Endpoint struct {
	ID          string           `json:"is"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string         `json:"mac"`
	Network     *Network
}

// 网络驱动
// 一个网络功能中的组件，不同的驱动对网络的创建、连接、销毁的策略不同
// 通过在创建网络时指定不同的网络驱动来定义使用那个驱动做网络的配置
type NetworkDriver interface {
	// 驱动名
	Name() string
	// 创建网络
	Create(subnet string, name string) (*Network, error)
	// 删除网络
	Delete(network Network) error
	// 连接容器网络端点到网络
	Connect(network *Network, endpoint *Endpoint) error
	// 从网络上移除容器网络端点
	Disconnect(network Network, endpoint *Endpoint) error
}

// 将这个网络的配置信息保存在文件系统中
func (nw *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}

	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorf("error：", err)
		return err
	}
	defer nwFile.Close()

	nwJson, err := json.Marshal(nw)
	if err != nil {
		logrus.Errorf("error：", err)
		return err
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		logrus.Errorf("error：", err)
		return err
	}
	return nil

}

// 从网络的配置目录中的文件读取到网络的配置
func (nw *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	defer nwConfigFile.Close()
	if err != nil {
		return err
	}
	nwJson := make([]byte, 4096)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		logrus.Errorf("Error load nw info", err)
		return err
	}
	return nil
}

// 从网络的配置目录中的删除配置文件
func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}

func CreateNetwork(driver, subnet, name string) error {
	// ParseCIDR 是 golang net的函数 功能是将网段的字符串转换成net.IPNet的对象
	_, cidr, _ := net.ParseCIDR(subnet)
	// 通过IPAM分配网关IP，获取到网段中第一个IP作为网关的IP
	gatewayIp, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}

	cidr.IP = gatewayIp
	// 通过指定的网络驱动创建网络，这里的drivers字典是哥哥网络驱动的实例字典，
	// 通过调用网络驱动的Create方法创建网络
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	return nw.dump(defaultNetworkPath)
}

```

#### 创建容器并连接网络

``` bash
mydocker run -it -p 80:80 --net testbridgenet xxxx
```

![img](./6.6)

## 容器地址分配

``` go
package network

import (
	"encoding/json"
	"net"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

//var ipamDefaultAllocatorPath = "/var/run/docker-ece/network/ipam/subnet.json"
var ipamDefaultAllocatorPath = "./subnet.json"

type IPAM struct {
	// 分配文件存放地址
	SubnetAllocatorPath string
	// 网段和位图算法的数组map key 是网段 value是分配的位图数组
	Subnets *map[string][]byte
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}
	subnetJson := make([]byte, 4096)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		logrus.Errorf("Error dump allocation info, %v", err)
		return err
	}
	return nil
}

func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(ipamConfigFileDir, 0644)
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath,
		os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}

	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}

	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return err
	}

	return nil
}

func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 存放网段中地址分配信息的数组
	ipam.Subnets = &map[string][]byte{}

	//从文件中加载已经分配的网段信息
	err = ipam.load()
	if err != nil {
		logrus.Errorf("Error dump allocation info, %v", err)
	}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	one, size := subnet.Mask.Size()

	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		// 初始化网段
		(*ipam.Subnets)[subnet.String()] = make([]byte, 1<<uint8(size-one))
	}
	for c := range (*ipam.Subnets)[subnet.String()] {
		if (*ipam.Subnets)[subnet.String()][c] == 0 {
			ipalloc := (*ipam.Subnets)[subnet.String()]
			ipalloc[c] = 1
			(*ipam.Subnets)[subnet.String()] = ipalloc
			ip = subnet.IP
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3] += 1
			break
		}
	}

	ipam.dump()
	return

}

// 释放地址
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string][]byte{}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		logrus.Errorf("Error dump allocation info, %v", err)
	}

	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = 0
	(*ipam.Subnets)[subnet.String()] = ipalloc

	ipam.dump()
	return nil
}

```

单元测试
``` go
package network

import (
	"fmt"
	"net"
	"testing"
)

func TestAllocate(t *testing.T) {
	// 在192.168.0.0/24网段内分配IP
	_, ipnet, _ := net.ParseCIDR("192.168.0.0/24")
	ip, _ := ipAllocator.Allocate(ipnet)
	t.Logf("alloc ip: %v", ip)
}

func TestRelease(t *testing.T) {
	// 在192.168.0.0/24网段中释放方才分配的IP
	ip, ipnet, _ := net.ParseCIDR("192.168.0.1/24")
	ipAllocator.Release(ipnet, &ip)
	fmt.Println(ip)
}

```

## 创建Bridge网络

![img](./6.9)
