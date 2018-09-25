<!-- TOC -->

- [client -> server 结构](#client---server-结构)
    - [指令](#指令)
        - [注册:](#注册)
        - [转发:](#转发)
- [server -> client 结构](#server---client-结构)

<!-- /TOC -->


# client -> server 结构

```
Cmd      string `json:"cmd"`
RoomID   string `json:"roomid"`
ClientID string `json:"clientid"`
Msg      string `json:"msg"`

```

## 指令
### 注册:
    * register
### 转发:
    * send

# server -> client 结构
```
Msg   string `json:"msg"`
Error string `json:"error"`
```
