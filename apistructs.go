package main

type idHolder struct {
	Id uint64 `json:"id"`
}

type User struct {
	ScreenName string `json:"screen_name"`
	Id         uint64 `json:"id"`
}

type FollowerList struct {
	NextCursor string  `json:"next_cursor_str"`
	Followers  []*User `json:"users"`
}

func (f *FollowerList) GetFollowerScreenNames() []string {
	ret := make([]string, len(f.Followers))
	for i, u := range f.Followers {
		ret[i] = u.ScreenName
	}
	return ret
}

type FollowerIDList struct {
	NextCursor string   `json:"next_cursor_str"`
	Followers  []uint64 `json:"ids"`
}

type TwitterErr struct {
	Msg    string
	Status int
}

func (err *TwitterErr) Error() string {
	return err.Msg
}

func NewTwitterErr(msg string, status int) *TwitterErr {
	return &TwitterErr{msg, status}
}
