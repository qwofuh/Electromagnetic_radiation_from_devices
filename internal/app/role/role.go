package role

type Role int

const (
	Guest Role = iota // 0
	Admin             // 1
	User              // 2
)
