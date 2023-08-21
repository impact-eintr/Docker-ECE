FROM centos
MAINTAINER <yixingwei4@gmail.com>
ENV MYPATH /usr/local
WORKDIR $MYPATH

# ADD 指令会自动解压
ADD apache-tomcat-9.0.79.tar.gz /usr/local/
ADD openjdk-8u43-linux-x64.tar.gz /usr/local/

# 添加软件源
RUN cd /etc/yum.repos.d/
RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*
RUN sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*

# 安装网络工具包
RUN yum -y install net-tools

#配置java与tomcat环境变量
ENV JAVA_HOME /usr/local/java-se-8u43-ri
ENV CLASSPATH $JAVA_HOME/lib/dt.jar:$JAVA_HOME/lib/tools.jar
ENV CATALINA_HOME /usr/local/apache-tomcat-9.0.79
ENV CATALINA_BASE /usr/local/apache-tomcat-9.0.79
ENV PATH $PATH:$JAVA_HOME/bin:$CATALINA_HOME/lib:$CATALINA_HOME/bin

EXPOSE 8080

CMD echo $MYPATH 
CMD echo "----------end--------" 
# 启动后执行tomcat
CMD /usr/local/apache-tomcat-9.0.79/bin/startup.sh && tail -F /usr/local/apache-tomcat-9.0.79/bin/logs/catalina.out
