package messengers

import (
	"context"
	"strconv"
	"strings"

	"github.com/treeder/gotils/v2"
)

func ParseCommand(ctx context.Context, text string) (string, []string) {
	t := text
	if strings.HasPrefix(t, "/") {
		t = t[1:]
	}
	tsplit := strings.Split(t, " ")
	tsplit = DeleteEmpty(tsplit)
	if len(tsplit) == 0 {
		return "", tsplit
	}
	cmd := strings.ToLower(tsplit[0])
	// if user clicks the help command, it comes in as /send@botname in Telegram
	cmdsplit := strings.Split(cmd, "@")
	cmd = cmdsplit[0]
	return cmd, tsplit
}

func DeleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// parses and logs errors
func ParseInt64(ctx context.Context, text string) int64 {
	mto, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		gotils.LogBeta(ctx, "error", "error converting %v to int64: %v", text, err)
		return mto
	}
	return mto
}
