# What is cotunnel?
[Cotunnel](https://www.cotunnel.com) is a service you can connect your to device's terminal, local webservers from everywhere. You don't need static ip, ddns or another service anymore. Cotunnel is your device's gateway to the world. 

> This repository includes cotunnel client. You can build and install cotunnel client from the source code following the instructions.

# Features

**Cloud Terminal**

You don't need to use SSH client anymore. You can connect to your device's terminal in cotunnel dashboard from eveywhere.

![cotunnel terminals page](https://cotunnel.s3.amazonaws.com/static/1.png)

**Tunnels**

Cotunnel creates a subdomain for your device and you can expose to the world in your local web servers. You can set cotunnel client should connect to the which local port in your cotunnel dashboard. Tunnel URLs support https connection too.

![cotunnel tunnels page](https://cotunnel.s3.amazonaws.com/static/3.png)

**Cross-platform** 

Cotunnel works with Linux, Mac, Windows also Raspberry Pi. 
> Terminals feature not working on Mac and Windows currently.

**Multiple Devices**

You can add your multiple devices to your cotunnel dashboard.

**Team Work**

You can add your workmates to your device's working group with privileges, then the users can edit or view your device. 

## How to build cotunnel client?

 - Install golang [https://golang.org/](https://golang.org/)
 - Install dep [https://github.com/golang/dep](https://github.com/golang/dep)
 >curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
 - Create your working folder named cotunnel this will be our gopath folder.
 >mkdir cotunnel
 
 >export GOPATH=(absolute path)/cotunnel
 - Create your source folder in cotunnel folder and clone this repository.
 >mkdir src && cd src
 
 >git clone gitlab.com/cotunnel/client
 
 >cd client
 - Get required dependencies using dep.
 >dep ensure
 - Build client, a file named "client" will be created.
 > go build

## How to install cotunnel client?

Before you start cotunnel client you need to the installation key. You can get installation key from cotunnel dashboard. Register to cotunnel dashboard. https://www.cotunnel.com/register Click add device button and it generates installation key.

- Start client with installation key.
> ./client --key yourinstallatonkey

- or start client and type installation key when it asks.
> ./client

If everything successfully works your device appears to your dashboard. You don't need to start cotunnel client with installation code anymore. Just start the cotunnel client.
> ./client
 
Do not delete cotunnel.key file. If you delete, you will lose your connection to your device in cotunnel dashboard. 
