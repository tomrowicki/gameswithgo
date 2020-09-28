package main

import (
	"gameswithgo/rpg/game"
	"gameswithgo/rpg/ui2d"
)

func main() {
	ui := &ui2d.UI2d{}
	game.Run(ui)
}
