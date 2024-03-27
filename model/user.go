package model

type User struct {
	ID       string  `chai:"id"`
	Username string  `chai:"username"`
	Password string  `chai:"password"`
	TTL      int64   `chai:"ttl"`
	Money    float64 `chai:"money"`
}

func (u User) Table() string {
	return "users"
}

func (u User) Public() map[string]any {
	return map[string]any{
		"ID":       u.ID,
		"Username": u.Username,
	}
}

var _ Model = &User{}
