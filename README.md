# prometheusgoproxy
A Prometheus pull proxy written in go

[![PatrickDch](https://img.shields.io/badge/PatrickDch-github-green)](https://github.com/PatrickDch)

# Introduction
Promproxy is a Prometheus pull proxy which you can install on a single or multiple bastion or jump hosts in order to pull eg. node_exporter in an nat`ed environment.

It is that easy, that you can not destroy it and it does not really have any smart output except for webrequests. --> KISS

No env-foo or automation, CI/CD needed. (Obviously you could use any of those)

Because go needs to be compiled by architecture if you want to have access to the source code, you can use the instructions like follows.

# Setup requirements
You need to have *golang* installed before you start building the code.

After that create the following requirements
```
mkdir -p /data/cert
mkdir -p /data/promproxy
openssl genrsa -out /data/cert/server.key 2048
openssl ecparam -genkey -name secp384r1 -out /data/cert/server.key
openssl req -x509 -nodes -sha256 -days 3650 -subj "/C=local/ST=local/O=local/CN=127.0.0.1" -addext "subjectAltName=IP:127.0.0.1" -key /data/cert/server.key -out /data/cert/server.crt;
```
like always we are going with a bit diffrent approach to TLS in this case *secp384r1* curves.

I highly suggest you use the systemd to control the Promproxy.
```
cp -v ./promproxy.service /etc/systemd/system/
```

If you are using a seperate machine to build your code, just seperate the requirements from building the code.

# Building the code
```
git clone https://github.com/PatrickDch/prometheusgoproxy.git
cd prometheusgoproxy
sudo go env -w GO111MODULE=auto
mkdir -p /opt/gobuild
cp -rv ./promproxy.go /opt/gobuild/
cd /opt/gobuild &&  sudo go get -v github.com/keep94/weblogs
cd /opt/gobuild &&  sudo go get -v github.com/gorilla/context
cd /opt/gobuild &&  sudo go build promproxy.go && sudo mv -vf promproxy /usr/local/bin/promproxy
```
Now you can start the Promproxy with systemd or as binary.
```
systemctl daemon-reload
systemctl restart promproxy.service
systemctl enable --now promproxy.service
```
Promproxy will be available on port :4445 but you could change that in the code.

For debbuging you could for example start netcat on the proxied end on which ever port you have choosen and check if the *checkme* hash is sent.

If it is not, meaning your Promproxie is not working.

# How is it working
To understand how exactly the proxy is working following a part of prometheus.yml

```
- job_name: 'server'
    scrape_interval: 11s
    proxy_url: https://promproxy:4445/
    static_configs:
    - targets: ['node_exporter:9091']
    basic_auth:
      username: 'username'
      password: 'password'
    scheme: http
    tls_config:
        insecure_skip_verify: true
```

First we need to work with *proxy_url* which is our Promproxy. (Listening on Port 4445)

Our *targets* are the actual node_export ip`s or FQDN, you could enter an array.

The server hosting Promproxy must reach ether the node_exporter ip or must be able to resolve the DNS name.

In my example I always use basic_auth to auth against the node_exporter !not the Promproxy!.

*scheme* needs to be *http*, it is kinda strange. There seems to be a bug in Prometheus while proxing requests the scheme is removed. 

But worry not the Promproxy is replacing it to *https* (It is good practice to expose your node_exporter with https only).

And thats that, that easy.
