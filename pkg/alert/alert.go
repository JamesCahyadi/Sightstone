package alert

import (
	"github.com/JamesCahyadi/Sightstone/assets"
	toast "github.com/electricbubble/go-toast"
)

func Send(message string) {
	_ = toast.Push(message,
		toast.WithTitle("Sighstone"),
		toast.WithIconRaw(assets.EmbededImage),
	)
}
