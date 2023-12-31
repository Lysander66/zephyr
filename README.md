# zephyr

Zephyr comes from Greek, which means "breeze", implying that this is a lightweight tool.

Zephyr 来自希腊语，表示“微风”，暗示着这是一个轻量级的工具。

[How to batch execute commands on multiple servers through BatchShell?](https://www.bilibili.com/video/BV1WN411r7gH)

[在多台服务器上批量运行命令](https://www.bilibili.com/video/BV1WN411r7gH)

`zephyr -h`

```
Usage of zephyr:
  -c string
    	command file path (default "cmd.txt")
  -l	run on local machine
  -p int
    	number of parallel hosts to process (default 100)
  -s string
    	ssh config file path (default "~/.ssh/config")
  -t int
    	timeout in seconds for each command (default 3600)
```

## Run on remote machine

e.g.
cmd.txt

```
// Host
host1
host2

// Command
curl -fsSL https://get.docker.com -o get-docker.sh
chmod +x ./get-docker.sh && ./get-docker.sh -y
```

## Run on local machine

`zephyr -l`

```
// Host
host1
host2

// Command
scp docker-compose ${HOST}:/usr/local/bin/docker-compose
```

this will execute

```sh
scp docker-compose host1:/usr/local/bin/docker-compose
scp docker-compose host2:/usr/local/bin/docker-compose
```
