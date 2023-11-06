#copyright IBM Corp. All Rights Reserved.
#
#SPDX-License-Identifier: Apache-2.0

From ubuntu:20.04
ARG DEBIAN_FRONTEND=noninteractive
ENV TZ=Asia/Shanghai

#RUN sed -i 's#http://archive.ubuntu.com/#http://mirrors.aliyun.com/#' /etc/apt/sources.list;
RUN apt-get update && apt-get install -y git libssl-dev
RUN apt-get update && apt-get install -y cmake gcc g++
RUN apt-get update && apt-get install -y make
RUN apt-get update && apt-get install -y unzip xsltproc doxygen graphviz
RUN apt-get update && apt-get install -y make clang++-7 libgmp-dev parallel
RUN apt-get update && apt-get install -y libtool libltdl-dev
RUN apt-get update && apt-get install -y libpcre3 libpcre3-dev
RUN apt-get update && apt-get install -y curl 
RUN apt-get update && apt-get install -y python3-setuptools 
RUN apt-get update && apt-get install -y python3-pip && pip3 --version
RUN apt-get update && apt-get install -y openssh-server


RUN python3 -m pip install --upgrade pip && pip3 --version
RUN pip3 install torch && pip3 install pandas && pip3 install torchvision
RUN pip3 install numpy

RUN pip3 install torchsummary  
RUN pip3 install scikit-learn && pip3 install Pillow


#node env
COPY node-v16.14.0-linux-x64.tar.xz /
RUN tar -C / -xvf node-v16.14.0-linux-x64.tar.xz
RUN ln -s /node-v16.14.0-linux-x64/bin/npm   /usr/local/bin/
RUN ln -s /node-v16.14.0-linux-x64/bin/node   /usr/local/bin/
RUN npm install -g npm@8.5.5 && npm install -g snarkjs@latest
RUN ln -s /node-v16.14.0-linux-x64/bin/snarkjs   /usr/local/bin/

RUN node --version && npm --version  
#golang env
COPY go1.20.1.linux-amd64.tar.gz /
RUN tar -C /usr/local -zxvf go1.20.1.linux-amd64.tar.gz
ENV GOROOT=/usr/local/go
ENV GOPATH=/root/go
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin
RUN go version

#java env 
#COPY jdk-8u162-linux-x64.tar.gz /
#RUN mkdir -p /usr/lib/jvm
#RUN tar -C /usr/lib/jvm -zxvf go1.17.8.linux-amd64.tar.gz
#RUN cd /usr/lib/jvm && ls
#ENV JAVA_HOME=/usr/lib/jvm/jdk1.8.0_162
#ENV JRE_HOME=${JAVA_HOME}/jre
#ENV CLASSPATH=.:${JAVA_HOME}/lib:${JRE_HOME}/lib
#ENV PATH=${JAVA_HOME}/bin:$PATH
#RUN java -version

COPY /payload/ML /ML/
COPY /payload/data /data/
COPY /payload/merkletest /merkletest/
RUN mkdir -p /chaincode/output /chaincode/input
RUN addgroup --gid 500 chaincode && useradd -d /home/chaincode -g chaincode chaincode
RUN chown -R chaincode:chaincode /chaincode
USER chaincode
USER root

#RUN go env -w GOPROXY=https://goproxy.cn,direct
#RUN go env -w GO111MODULE=on
