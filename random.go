package messengers

import (
	"math/rand"

	"github.com/rs/xid"
)

// ChooseRandom chooses a random element from slice
func ChooseRandom(sl []string) string {
	if sl == nil {
		return ""
	}
	return sl[rand.Intn(len(sl))]
}

func GenToken() string {
	guid := xid.New()
	return guid.String()
}
