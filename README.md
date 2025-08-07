为`handy-sshd`添加密钥支持，再也不用密码登录了！
使用方法:
1. 从Actions/CI下载最新的可执行文件
2. 使用 `-k <authorized_keys_path>`指定存储公钥的路径 (如果同时使用`-u <user>:<password>`，则允许同时使用用户名和密码登录)
3. 使用正确的公私钥连接

来源: [nwtgck/handy-sshd PR #34](https://github.com/nwtgck/handy-sshd/pull/34/commits/8c71f5de349d05b0f5025237cad3314902a01cbe)

该项目在 [wujinjun-MC/simpfun-better-start-script](https://github.com/wujinjun-MC/simpfun-better-start-script/) 也有使用


以下是原README:

# handy-sshd
[![CI](https://github.com/nwtgck/handy-sshd/actions/workflows/ci.yml/badge.svg)](https://github.com/nwtgck/handy-sshd/actions/workflows/ci.yml)

Portable SSH Server

## Install on Ubuntu/Debian

```bash
wget https://github.com/nwtgck/handy-sshd/releases/download/v0.4.2/handy-sshd-0.4.2-linux-amd64.deb
sudo dpkg -i handy-sshd-0.4.2-linux-amd64.deb 
```

## Install on Mac

```bash
brew install nwtgck/handy-sshd/handy-sshd
```

Get more executables in [the releases](https://github.com/nwtgck/handy-sshd/releases).

## Examples

```bash
# Listen on 2222 and accept user name "john" with password "mypass"
handy-sshd -p 2222 -u john:mypass
```

```bash
# Listen on 2222 and accept user name "john" without password
handy-sshd -p 2222 -u john:
```

```bash
# Listen on 2222 and accept users "john" and "alice" without password
handy-sshd -p 2222 -u john: -u alice:
```

```bash
# Listen on unix domain socket
handy-sshd --unix-socket /tmp/my-unix-socket -u john:
```

## Features
An SSH client can use
* Shell/Interactive shell
* Local port forwarding (ssh -L)
* Remote port forwarding (ssh -R)
* [SOCKS proxy](https://wikipedia.org/wiki/SOCKS) (dynamic port forwarding)
* SFTP
* [SSHFS](https://wikipedia.org/wiki/SSHFS)
* Unix domain socket (local/remote port forwarding)

All features are enabled by default. You can allow only some of them using permission flags.

## Permissions
There are several permissions:
* --allow-direct-streamlocal
* --allow-direct-tcpip
* --allow-execute
* --allow-sftp
* --allow-streamlocal-forward
* --allow-tcpip-forward

**All permissions are allowed when nothing is specified.** The log shows "allowed: " and "NOT allowed: " permissions as follows:

```console
$ handy-sshd -u "john:"
2023/08/11 11:40:44 INFO listening on :2222...
2023/08/11 11:40:44 INFO allowed: "tcpip-forward", "direct-tcpip", "execute", "sftp", "streamlocal-forward", "direct-streamlocal"
2023/08/11 11:40:44 INFO NOT allowed: none
```

For example, specifying `--allow-direct-tcpip` and `--allow-execute` allows only them:

```console
$ handy-sshd -u "john:" --allow-direct-tcpip --allow-execute
2023/08/11 11:41:03 INFO listening on :2222...
2023/08/11 11:41:03 INFO allowed: "direct-tcpip", "execute"
2023/08/11 11:41:03 INFO NOT allowed: "tcpip-forward", "sftp", "streamlocal-forward", "direct-streamlocal"
```

## --help

```
Portable SSH server

Usage:
  handy-sshd [flags]

Examples:
# Listen on 2222 and accept user name "john" with password "mypass"
handy-sshd -u john:mypass

# Listen on 22 and accept the user without password
handy-sshd -p 22 -u john:

Permissions:
All permissions are allowed by default.
For example, specifying --allow-direct-tcpip and --allow-execute allows only them.

Flags:
      --allow-direct-streamlocal    client can use Unix domain socket local forwarding (ssh -L)
      --allow-direct-tcpip          client can use local forwarding (ssh -L) and SOCKS proxy (ssh -D)
      --allow-execute               client can use shell/interactive shell
      --allow-sftp                  client can use SFTP and SSHFS
      --allow-streamlocal-forward   client can use Unix domain socket remote forwarding (ssh -R)
      --allow-tcpip-forward         client can use remote forwarding (ssh -R)
  -h, --help                        help for handy-sshd
      --host string                 SSH server host to listen (e.g. 127.0.0.1)
  -p, --port uint16                 port to listen (default 2222)
      --shell string                Shell
      --unix-socket string          Unix domain socket to listen
  -u, --user stringArray            SSH user name (e.g. "john:mypass")
  -v, --version                     show version
```
