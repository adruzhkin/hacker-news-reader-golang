package model

// User holds a username and their associated comment count.
type User struct {
	Name  string
	Count int
}

// UserList is a sortable slice of User, ordered by descending count with alphabetical tiebreak.
type UserList []User

// Len implements sort.Interface.
func (l UserList) Len() int {
	return len(l)
}

// Swap implements sort.Interface.
func (l UserList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// Less implements sort.Interface. It sorts by descending count, with alphabetical tiebreak.
func (l UserList) Less(i, j int) bool {
	if l[i].Count == l[j].Count {
		return l[i].Name < l[j].Name
	}
	return l[i].Count > l[j].Count
}
