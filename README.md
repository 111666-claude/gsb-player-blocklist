# player-blocklist

玩家拉黑名单：名额上限、与好友互斥、请求幂等，状态存 JSON 文件，用 `block` / `unblock` / `list` 子命令管理，
只用 Go 标准库。

```
go run ./cmd/player-blocklist --sample=cap
go run ./cmd/player-blocklist --sample=friend
go run ./cmd/player-blocklist --sample=idem
go run ./cmd/player-blocklist --sample=round
go run ./cmd/player-blocklist block r1 p1 spammer --state state.json
go run ./cmd/player-blocklist list p1 --state state.json
go test ./...
go vet ./...
```

## 口径（README 为准）

- **名额上限**：每个玩家拉黑人数不超过 `max`，超出返回 `ErrTooMany` 且名单不变。
- **好友互斥**：对方在自己好友名单里时返回 `ErrFriend` 且名单不变（好友是双向的）。
- **幂等**：同一个 `req_id` 只处理一次，重复提交返回 `ErrReplayed`。
- **解除**：`Unblock` 只在名单里有这个人时返回 true 并移除。
- **往返一致**：`Save` 写出的 JSON 用 `Load` 读回来，名额上限、拉黑名单、好友关系与已处理请求都必须一致。
- **代价**：查询按玩家索引，摊还 O(1)；10 万玩家每秒 10 万次拉黑与查询。

## 输出契约（不改格式）

```
ok=..
second=..
blocked=..
```

## 目录

```
blocklist/list.go              上限、互斥、幂等与 JSON 往返
cmd/player-blocklist/main.go   子命令入口
blocklist/list_test.go         单元用例
```
