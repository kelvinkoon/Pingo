# Pingo

An implementation of a Ping CLI application as described by Cloudflare's Summer 2020 Internship Application [take-home assignment](https://github.com/cloudflare-internship-2020/internship-application-systems). Features IPv4 and IPv6 support.

## Examples

Note: `sudo` is required due to socket usage.
Use `./run_demo.sh` to run the default ping (IPv4 to `github.com`).

### IPv4 Ping

```
sudo go run src/client/main.go -host=gobyexample.com -ip=4

INFO[0000] Reply from 99.84.238.127: bytes=32 time=27ms
INFO[0001] Reply from 99.84.238.127: bytes=32 time=17ms
```

### IPv6 Ping

```
sudo go run src/client/main.go -host=google.com -ip=6

INFO[0000] Reply from 2607:f8b0:4005:80f::200e: bytes=32 time=15ms
INFO[0001] Reply from 2607:f8b0:4005:80f::200e: bytes=32 time=21ms
```
