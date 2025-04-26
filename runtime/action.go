package runtime

type ActionGroup int

const (
	Camera ActionGroup = iota
	Home
	Record
	Chat
)

type Action struct {
	Name  string
	Title string
	Icon  string
	Group ActionGroup
}
