package main

import (
	"gameswithgo/rpg/game"
	"gameswithgo/rpg/ui2d"
)

func main() {
	//numWindows := 1 // does not work with > 1
	//
	//rpgGame := game.NewGame(numWindows, "rpg/game/maps/level1.map")
	//
	//for i := 0; i < numWindows; i++ {
	//	go func(i int) {
	//		runtime.LockOSThread() // OSes require main app thread to stay associated with one system thread
	//		ui := ui2d.NewUI(rpgGame.InputChan, rpgGame.LevelChans[i])
	//		ui.Run()
	//	}(i)
	//}
	//
	//rpgGame.Run()

	//rpg := game.NewGame(1, "rpg/game/maps/level1.map")
	rpg := game.NewGame(1)

	go func() { rpg.Run() }()
	ui := ui2d.NewUI(rpg.InputChan, rpg.LevelChans[0])
	ui.Run()
}
