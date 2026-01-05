# UDPping
ping with UDP protocol 🛠

# How it looks like
```
root@bitch:~# ./udpping 44.55.66.77 4000
UDPping 44.55.66.77 via port 4000 with 64 bytes of payload
Reply from 44.55.66.77 seq=0 time=138.357 ms
Reply from 44.55.66.77 seq=1 time=128.062 ms
Request timed out
Reply from 44.55.66.77 seq=3 time=136.370 ms
Reply from 44.55.66.77 seq=4 time=140.743 ms
Request timed out
Reply from 44.55.66.77 seq=6 time=143.438 ms
Reply from 44.55.66.77 seq=7 time=142.684 ms
Reply from 44.55.66.77 seq=8 time=138.871 ms
Reply from 44.55.66.77 seq=9 time=138.990 ms
^C
--- ping statistics ---
10 packets transmitted, 8 received, 20.00% packet loss
rtt min/avg/max = 128.06/138.44/143.44 ms
```

# Getting Started

### Step 1

Set up a udp echo server at the host you want to ping.

There are many ways of doing this, my favourite way is:

```
socat -v UDP-LISTEN:4000,fork PIPE
```

For IPv6:

```
socat -v UDP6-LISTEN:4000,fork PIPE
```

Now a echo server is listening at port 4000.

###### Note
If you dont have socat, use `apt install socat` or `yum install socat`, you will get it.

### Step 2

Ping you server.

Assume `44.55.66.77` is the IP of your server.

```
./udpping 44.55.66.77 4000
```

Done!

Now UDPping will generate outputs as a normal ping, but the protocol used is `UDP` instead of `ICMP`.

# Advanced Usage
```
root@bitch:~# ./udpping
usage:
  udpping <dest_ip> <dest_port>
  udpping <dest_ip> <dest_port> [options]

options:
  -len        Payload length in bytes (default: 64)
  -interval   Interval between packets in milliseconds (default: 1000)
  -4          Force IPv4
  -6          Force IPv6

examples:
  udpping 44.55.66.77 4000
  udpping fe80::5400:ff:aabb:ccdd 4000
  udpping example.com 4000
  udpping example.com 4000 -4
  udpping 44.55.66.77 4000 -len=400 -interval=2000
```

# Features

- **IPv4 and IPv6 Support**: Automatically detects and supports both IPv4 and IPv6 addresses
- **Domain Name Resolution**: Supports domain names with automatic DNS resolution
- **IP Version Control**: Use -4 or -6 flags to force IPv4 or IPv6 resolution
- **Automatic Reconnection**: Automatically reconnects if connection fails
- **Configurable Payload**: Customize payload size with `-len` parameter
- **Configurable Interval**: Set ping interval and timeout with `-interval` parameter
- **Statistics**: Shows packet loss and RTT statistics on exit (Ctrl+C)

# Building from Source

```bash
# Build for current platform
go build -o udpping main.go

# Build for Linux amd64
GOOS=linux GOARCH=amd64 go build -o udpping main.go

# Build for Windows amd64
GOOS=windows GOARCH=amd64 go build -o udpping.exe main.go
```
