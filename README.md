# zephyr

Zephyr comes from Greek, which means "breeze", implying that this is a lightweight tool.

Zephyr来自希腊语，表示“微风”，暗示着这是一个轻量级的工具。

zephyr -h
```
Usage of zephyr:
  -c string
        command file path (default "cmd.txt")
  -p int
        number of parallel hosts to process (default 10)
  -s string
        ssh config file path (default "~/.ssh/config")
  -t int
        timeout in seconds for each command (default 30)
```

e.g.
cmd.txt
```
// Host
host1
host2

// Command
cd /srv
date +"%Y/%m/%d %H:%M:%S"
sleep 2
date +"%Y/%m/%d %H:%M:%S"
ls
```
