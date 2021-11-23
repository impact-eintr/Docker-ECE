# Docker-ECE
支持cgroup2和overlay2的容器运行时

### 安装

``` bash
gh repo clone impact-eintr/Docker-ECE

cd Docker-ECE

go build

sudo ./Docker-ECE
```

### 简单使用

``` bash
sudo ./Docker-ECE run -it /bin/sh 

INFO[0000] Docker-ECE is a simple container runtime implementation
/bin/sh
INFO[0000] Docker-ECE is a simple container runtime implementation
{"level":"info","msg":"init come on","time":"2021-11-23T11:18:12+08:00"}
{"level":"info","msg":"command ","time":"2021-11-23T11:18:12+08:00"}
{"level":"info","msg":"read parent pipe cmd","time":"2021-11-23T11:18:12+08:00"}
{"level":"info","msg":"command all is /bin/sh","time":"2021-11-23T11:18:12+08:00"}
{"level":"info","msg":"receive /bin/sh","time":"2021-11-23T11:18:12+08:00"}
{"level":"info","msg":"Current location is [/var/lib/docker-ece/HEZDENJXG43TQMBT/merge]","time":"2021-11-23T11:18:12+08:00"}
{"level":"info","msg":"now change dir to root","time":"2021-11-23T11:18:12+08:00"}
/ # export PATH=/bin
/ # ls
bin   dev   etc   home  proc  root  sys   tmp   usr   var

```

### 使用已有的镜像

这里就先使用docker导出的镜像了

``` bash
docker run -it ubuntu /bin/bash 

docker ps
CONTAINER ID   IMAGE       COMMAND                  CREATED        STATUS                   PORTS     NAMES
de8196e01926   ubuntu      "/bin/bash"              2 weeks ago    Exited (0) 2 weeks ago             beautiful_hoover

docker export -o myubuntu.tar de8196e01926

# 将镜像放到指定位置 Docker-ECE/Images 否则找不到

mv myubuntu.tar Docker-ECE/Images
```

注意名字要对上 不需要加`.tar` 直接名字就行

``` bash
sudo ./Docker-ECE run -it --image myubuntu /bin/bash

## 以下是正确进入一个ubuntu容器的样子

INFO[0000] Docker-ECE is a simple container runtime implementation
/bin/bash
INFO[0000] Docker-ECE is a simple container runtime implementation
{"level":"info","msg":"init come on","time":"2021-11-23T10:49:57+08:00"}
{"level":"info","msg":"command ","time":"2021-11-23T10:49:57+08:00"}
{"level":"info","msg":"read parent pipe cmd","time":"2021-11-23T10:49:57+08:00"}
cgroup2
{"level":"info","msg":"command all is /bin/bash","time":"2021-11-23T10:49:57+08:00"}
{"level":"info","msg":"receive /bin/bash","time":"2021-11-23T10:49:57+08:00"}
{"level":"info","msg":"Current location is [/var/lib/docker-ece/GMZTEOJVGIZDCMBZ/merge]","time":"2021-11-23T10:49:57+08:00"}
{"level":"info","msg":"now change dir to root","time":"2021-11-23T10:49:57+08:00"}
root@Code01:/# ls
bin   dev  home  lib32  libx32  mnt  proc  run   srv  tmp  var
boot  etc  lib   lib64  media   opt  root  sbin  sys  usr
root@Code01:/# ps
    PID TTY          TIME CMD
      1 ?        00:00:00 bash
     11 ?        00:00:00 ps


```

### 容器联网

先看一下本地的DNS

``` bash
cat /etc/resolv.conf
```

``` bash
> sudo ./Docker-ECE run -it --net test /bin/sh
[sudo] eintr 的密码：
INFO[0000] Docker-ECE is a simple container runtime implementation
/bin/sh
INFO[0000] Docker-ECE is a simple container runtime implementation
{"level":"info","msg":"init come on","time":"2021-11-23T11:38:04+08:00"}
{"level":"info","msg":"command ","time":"2021-11-23T11:38:04+08:00"}
{"level":"info","msg":"read parent pipe cmd","time":"2021-11-23T11:38:04+08:00"}
{"level":"info","msg":"command all is /bin/sh","time":"2021-11-23T11:38:04+08:00"}
{"level":"info","msg":"receive /bin/sh","time":"2021-11-23T11:38:04+08:00"}
{"level":"info","msg":"Current location is [/var/lib/docker-ece/GIYTQNZVGU2TMMZV/merge]","time":"2021-11-23T11:38:04+08:00"}
{"level":"info","msg":"now change dir to root","time":"2021-11-23T11:38:04+08:00"}
/ # export PATH=/bin
/ # echo "nameserver 61.139.2.69" > /etc/resolv.conf // 把查出来的DNS添进去
/ # ping www.baidu.com
PING www.baidu.com (14.215.177.38): 56 data bytes
64 bytes from 14.215.177.38: seq=0 ttl=53 time=36.101 ms
64 bytes from 14.215.177.38: seq=1 ttl=53 time=34.067 ms
64 bytes from 14.215.177.38: seq=2 ttl=53 time=34.307 ms
64 bytes from 14.215.177.38: seq=3 ttl=53 time=33.259 ms
^C
--- www.baidu.com ping statistics ---
4 packets transmitted, 4 packets received, 0% packet loss
round-trip min/avg/max = 33.259/34.433/36.101 ms
/ #

```

# 《自己动手写docker》的学习笔记
## 基础知识
<https://github.com/impact-eintr/Docker-ECE/tree/main/note/basic/README.md>

## 容器网络
<https://github.com/impact-eintr/Docker-ECE/blob/main/note/network/README.md>

