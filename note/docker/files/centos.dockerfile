FROM centos
MAINTAINER <yixingwei4@gmail.com>
ENV MYPATH /usr/local
WORKDIR $MYPATH

# 添加软件源
RUN cd /etc/yum.repos.d/
RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*
RUN sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*

# 安装网络工具包
RUN yum -y install net-tools

EXPOSE 80

CMD echo $MYPATH 
CMD echo "----------end--------" 
CMD /bin/bash
