# 构建镜像

## busybox
首先使用一个最精简的镜像——busybox busybox 个集合了非常多 山叫IX 工具的箱子，
它可以提供非常多在UNIX 环境下经常使用的命令，可以说 busybox 提供了一个非常完整而且小巧的系统。本小节会先使用它来作为第 个容器内运行的文件系统。

获得 busybox 文件系统的 rootfs 很简单，可以使用 docker export 将一个镜像打成一个tar包。

首先做如下操作。

``` bssh

docker pull busybox 

docker run -d busybox top -b 

docker export - o busybox.tar 835360ffl6b8 （容器 ID)

tar -xvf busybox . tar -C busybox/
```

## pivot_root
`pivot_root`是一个系统调用，主要功能是去改变当前的root文件系统。pivot_root可以将当前进程的root文件系统移动到put_old文件夹中，然后使new_root成为新的root文件系统。
new_root 和 put_old必须不能存在当前root的同一个文件系统中。pivot_root和chroot的主要区别是，pivot_root是把整个系统切换到一个新的root目录，而移除对之前root文件系统的依赖，这样你就能够umount原先的root文件系统。而chroot是针对某个进程，系统的其他部分依旧运行于老的root目录中

``` go

```


