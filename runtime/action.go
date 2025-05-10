package runtime

type ActionGroup int

const (
	Camera ActionGroup = iota
	Home
	Record
	Chat
)

type Action struct {
	Name  string      `json:"name"`
	Title string      `json:"title"`
	Icon  string      `json:"icon"`
	Group ActionGroup `json:"group"`
}
