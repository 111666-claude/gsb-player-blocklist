package blocklist

import "testing"

func TestBlockThenListed(t *testing.T) {
	list := New(3)
	if err := list.Block("r1", "p1", "p2"); err != nil {
		t.Fatal(err)
	}
	if got := list.Blocked("p1"); len(got) != 1 || got[0] != "p2" {
		t.Fatalf("拉黑名单应该是 [p2]，得到 %v", got)
	}
}

func TestUnblockRemoves(t *testing.T) {
	list := New(3)
	list.Block("r1", "p1", "p2")
	if !list.Unblock("p1", "p2") {
		t.Fatal("解除拉黑应该返回 true")
	}
	if len(list.Blocked("p1")) != 0 {
		t.Fatal("解除之后名单应该是空的")
	}
}

func TestBlockUnknownTargetIsAllowed(t *testing.T) {
	if err := New(3).Block("r1", "p1", "stranger"); err != nil {
		t.Fatalf("拉黑陌生人应该成功，得到 %v", err)
	}
}
