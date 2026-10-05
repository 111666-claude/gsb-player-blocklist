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

// Block 拉黑一个玩家：reqID 幂等，好友互斥，名额受限，任一不满足则名单不变。
func (l *List) Block(reqID, player, target string) error {
	if l.seen[reqID] {
		return ErrReplayed
	}
	if l.friends[player][target] {
		return ErrFriend
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

// Save 序列化台账：名额上限、拉黑名单、好友关系与已处理请求全部落盘。
func (l *List) Save() []byte {
	body, _ := json.Marshal(map[string]any{
		"max":     l.max,
		"blocked": l.blocked,
		"friends": l.friends,
		"seen":    l.seen,
	})
	return body
}

// Load 从 JSON 恢复台账。
func (l *List) Load(data []byte) error {
	payload := struct {
		Max     int                        `json:"max"`
		Blocked map[string][]string        `json:"blocked"`
		Friends map[string]map[string]bool `json:"friends"`
		Seen    map[string]bool            `json:"seen"`
	}{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	l.max = payload.Max
	if payload.Blocked != nil {
		l.blocked = payload.Blocked
	}
	if payload.Friends != nil {
		l.friends = payload.Friends
	}
	if payload.Seen != nil {
		l.seen = payload.Seen
	}
	return nil
}
