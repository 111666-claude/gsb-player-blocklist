// Command player-blocklist 用子命令管理拉黑名单，状态存 JSON 文件。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"example.com/blocklist/blocklist"
)

func load(path string, max int) (*blocklist.List, error) {
	list := blocklist.New(max)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return list, nil
		}
		return nil, err
	}
	if err := list.Load(data); err != nil {
		return nil, err
	}
	return list, nil
}

func save(path string, list *blocklist.List) error {
	return os.WriteFile(path, list.Save(), 0o644)
}

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("player-blocklist", flag.ContinueOnError)
	flags.SetOutput(stderr)
	state := flags.String("state", "state.json", "状态文件")
	max := flags.Int("max", 50, "每人拉黑上限")
	sample := flags.String("sample", "", "cap / friend / idem / round")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *sample != "" {
		return runSample(*sample, *max, stdout, stderr)
	}
	rest := flags.Args()
	if len(rest) < 2 {
		fmt.Fprintln(stderr, "需要子命令 block|unblock|list")
		return 2
	}
	list, err := load(*state, *max)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	switch rest[0] {
	case "block":
		if len(rest) < 3 {
			fmt.Fprintln(stderr, "用法：block <req> <player> <target>")
			return 2
		}
		fmt.Fprintf(stdout, "ok=%v\n", list.Block(rest[1], rest[2], rest[3]) == nil)
	case "unblock":
		fmt.Fprintf(stdout, "ok=%v\n", list.Unblock(rest[1], rest[2]))
	case "list":
		fmt.Fprintf(stdout, "blocked=%s\n", strings.Join(list.Blocked(rest[1]), ","))
	default:
		fmt.Fprintln(stderr, "未知子命令")
		return 2
	}
	if err := save(*state, list); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func runSample(name string, max int, stdout, stderr io.Writer) int {
	switch name {
	case "cap":
		list := blocklist.New(1)
		list.Block("r1", "p1", "a")
		fmt.Fprintf(stdout, "second=%v\n", list.Block("r2", "p1", "b") == nil)
	case "friend":
		list := blocklist.New(10)
		list.AddFriend("p1", "a")
		fmt.Fprintf(stdout, "ok=%v\n", list.Block("r1", "p1", "a") == nil)
	case "idem":
		list := blocklist.New(10)
		list.Block("r1", "p1", "a")
		fmt.Fprintf(stdout, "second=%v\n", list.Block("r1", "p1", "b") == nil)
	case "round":
		list := blocklist.New(10)
		list.AddFriend("p1", "b")
		data := list.Save()
		restored := blocklist.New(10)
		if err := restored.Load(data); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "ok=%v\n", restored.Block("r1", "p1", "b") == nil)
	default:
		fmt.Fprintln(stderr, "需要 --sample cap|friend|idem|round")
		return 2
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
