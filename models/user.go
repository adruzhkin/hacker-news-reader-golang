package models

type User struct {
	Name  string
	Count int
}

type UserList []User

func (l UserList) Len() int {
	return len(l)
}

func (l UserList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l UserList) Less(i, j int) bool {
	if l[i].Count == l[j].Count {
		return l[i].Name < l[j].Name
	}
	return l[i].Count > l[j].Count
}
