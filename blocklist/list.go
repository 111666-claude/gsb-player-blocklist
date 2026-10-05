// Package blocklist 是玩家拉黑名单：名额上限、与好友互斥、请求幂等与 JSON 状态往返。
package blocklist

import (
	"encoding/json"
	"errors"
	"sort"
)

// ErrTooMany 拉黑名额已满。
var ErrTooMany = errors.New("拉黑名额已满")

// ErrFriend 对方是好友，不能拉黑。
var ErrFriend = errors.New("不能拉黑好友")

// ErrReplayed 这个请求已经处理过。
var ErrReplayed = errors.New("请求已处理")

// List 是拉黑台账。
type List struct {
	max     int
	blocked map[string][]string
	friends map[string]map[string]bool
	seen    map[string]bool
}

// New 建台账：max 是每人最多能拉黑的人数。
func New(max int) *List {
	return &List{max: max, blocked: map[string][]string{}, friends: map[string]map[string]bool{}, seen: map[string]bool{}}
}

// AddFriend 登记一对好友（双向）。
func (l *List) AddFriend(a, b string) {
	for _, pair := range [][2]string{{a, b}, {b, a}} {
		if l.friends[pair[0]] == nil {
			l.friends[pair[0]] = map[string]bool{}
		}
		l.friends[pair[0]][pair[1]] = true
	}
}

// Block 拉黑一个玩家：reqID 重复返回 ErrReplayed；对方是好友返回 ErrFriend；
// 超过名额上限返回 ErrTooMany。任何失败都不改变台账。
func (l *List) Block(reqID, player, target string) error {
	if l.seen[reqID] {
		return ErrReplayed
	}
	if l.friends[player][target] {
		return ErrFriend
	}
	for _, name := range l.blocked[player] {
		if name == target {
			l.seen[reqID] = true
			return nil
		}
	}
	if len(l.blocked[player]) >= l.max {
		return ErrTooMany
	}
	l.blocked[player] = append(l.blocked[player], target)
	l.seen[reqID] = true
	return nil
}

// Unblock 解除拉黑，返回是否真的解除了。
func (l *List) Unblock(player, target string) bool {
	items := l.blocked[player]
	for index, name := range items {
		if name == target {
			l.blocked[player] = append(items[:index], items[index+1:]...)
			return true
		}
	}
	return false
}

// Blocked 返回该玩家的拉黑名单（升序）。
func (l *List) Blocked(player string) []string {
	out := append([]string{}, l.blocked[player]...)
	sort.Strings(out)
	return out
}

// snapshot 是台账的 JSON 表示，名额上限、拉黑名单、好友关系与已处理请求全部落盘。
type snapshot struct {
	Max     int                 `json:"max"`
	Blocked map[string][]string `json:"blocked"`
	Friends map[string][]string `json:"friends"`
	Seen    []string            `json:"seen"`
}

// Save 序列化台账。
func (l *List) Save() []byte {
	friends := map[string][]string{}
	for player, set := range l.friends {
		names := make([]string, 0, len(set))
		for name := range set {
			names = append(names, name)
		}
		sort.Strings(names)
		friends[player] = names
	}
	seen := make([]string, 0, len(l.seen))
	for reqID := range l.seen {
		seen = append(seen, reqID)
	}
	sort.Strings(seen)
	body, _ := json.Marshal(snapshot{
		Max:     l.max,
		Blocked: l.blocked,
		Friends: friends,
		Seen:    seen,
	})
	return body
}

// Load 从 JSON 恢复台账。
func (l *List) Load(data []byte) error {
	payload := snapshot{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	l.max = payload.Max
	l.blocked = map[string][]string{}
	if payload.Blocked != nil {
		l.blocked = payload.Blocked
	}
	l.friends = map[string]map[string]bool{}
	for player, names := range payload.Friends {
		set := map[string]bool{}
		for _, name := range names {
			set[name] = true
		}
		l.friends[player] = set
	}
	l.seen = map[string]bool{}
	for _, reqID := range payload.Seen {
		l.seen[reqID] = true
	}
	return nil
}
