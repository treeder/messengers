package messengers

import "github.com/treeder/messengers/ui"

type ResponseMsg struct {
	Text  string     `json:"text"`
	Cards []*ui.Card `json:"cards,omitempty"`
}
