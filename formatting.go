package messengers

import (
	"fmt"
	"strings"
)

type FormatStr int

const (
	Regular FormatStr = iota
	Bold
	Italic
)

// note: this was for telegram specifically, might need to change for other messengers
var replacer = strings.NewReplacer("_", "\\_", "*", "\\*", "[", "\\[", "`", "\\`")

func MarkdownEscape(s string) string {
	return replacer.Replace(s)
}

func MarkdownEscapeErr(err error) string {
	return MarkdownEscape(err.Error())
}

func SprintUsername(msg IncomingMessage) string {
	if msg.IsPrivate() {
		return ""
	}
	return fmt.Sprintf("%v ", msg.Mention())
}
