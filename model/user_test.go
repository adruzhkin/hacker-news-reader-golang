package model

import (
	"sort"
	"testing"
)

func TestUserListSort_DescendingByCount(t *testing.T) {
	list := UserList{
		{Name: "alice", Count: 1},
		{Name: "bob", Count: 3},
		{Name: "carol", Count: 2},
	}
	sort.Sort(list)

	want := []int{3, 2, 1}
	for i, w := range want {
		if list[i].Count != w {
			t.Errorf("index %d: got count %d, want %d", i, list[i].Count, w)
		}
	}
}

func TestUserListSort_AlphabeticalTiebreak(t *testing.T) {
	list := UserList{
		{Name: "carol", Count: 5},
		{Name: "alice", Count: 5},
		{Name: "bob", Count: 5},
	}
	sort.Sort(list)

	want := []string{"alice", "bob", "carol"}
	for i, w := range want {
		if list[i].Name != w {
			t.Errorf("index %d: got name %q, want %q", i, list[i].Name, w)
		}
	}
}

func TestUserListSort_SingleElement(t *testing.T) {
	list := UserList{{Name: "alice", Count: 1}}
	sort.Sort(list)

	if list[0].Name != "alice" || list[0].Count != 1 {
		t.Errorf("unexpected result: %+v", list[0])
	}
}

func TestUserListSort_Empty(t *testing.T) {
	list := UserList{}
	sort.Sort(list) // should not panic
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d elements", len(list))
	}
}

func TestUserListSort_MixedTiebreakAndCount(t *testing.T) {
	list := UserList{
		{Name: "dave", Count: 2},
		{Name: "alice", Count: 5},
		{Name: "bob", Count: 5},
		{Name: "carol", Count: 2},
		{Name: "eve", Count: 10},
	}
	sort.Sort(list)

	want := []struct {
		name  string
		count int
	}{
		{"eve", 10},
		{"alice", 5},
		{"bob", 5},
		{"carol", 2},
		{"dave", 2},
	}

	for i, w := range want {
		if list[i].Name != w.name || list[i].Count != w.count {
			t.Errorf("index %d: got {%s, %d}, want {%s, %d}",
				i, list[i].Name, list[i].Count, w.name, w.count)
		}
	}
}
