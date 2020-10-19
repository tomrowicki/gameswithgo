package game

import "fmt"

type Monster struct {
	Character
}

func (m *Monster) Kill(level *Level) {
	delete(level.Monsters, m.Pos)
	groundItems := level.Items[m.Pos]
	for _, item := range m.Items {
		item.Pos = m.Pos
		groundItems = append(groundItems, item)
	}
	level.Items[m.Pos] = groundItems
}

func NewRat(p Pos) *Monster {
	//return &Monster{Pos:p, Rune:'R', Name: "Rat", Hitpoints:5, Strength:5, Speed:1.5, ActionPoints:0.0}
	return &Monster{
		Character{
			Entity: Entity{
				Pos:  p,
				Name: "Rat",
				Rune: 'R',
			},
			Hitpoints:    50,
			Strength:     1,
			Speed:        2.0,
			ActionPoints: 0.0,
			SightRange: 10,
			Items:[]*Item{NewHelmet(p)},
		},
	}
}

func NewSpider(p Pos) *Monster {
	//return &Monster{p, 'S', "Spider", 10, 10, 1.0, .0}
	return &Monster{Character{
		Entity: Entity{
			Pos:  p,
			Name: "Spider",
			Rune: 'S',
		},
		Hitpoints:    100,
		Strength:     1,
		Speed:        1.0,
		ActionPoints: 0.0,
		SightRange: 10,
		Items: []*Item{NewSword(p)},
	}}
}

func (m *Monster) Update(level *Level) {
	m.ActionPoints += m.Speed
	playerPos := level.Player.Pos

	apInt := int(m.ActionPoints)
	positions := level.astar(m.Pos, playerPos)

	if len(positions) == 0 {
		m.Pass()
		return
	}

	moveIndex := 1
	for i := 0; i < apInt; i++ {
		if moveIndex < len(positions) {
			m.Move(positions[moveIndex], level)
			moveIndex++
			m.ActionPoints--
		}
	}
}

func (m *Monster) Pass() {
	m.ActionPoints -= m.Speed
}

func (m *Monster) Move(to Pos, level *Level) {
	_, exists := level.Monsters[to]

	// TODO check if tile being moved to is valid
	if !exists && to != level.Player.Pos {
		delete(level.Monsters, m.Pos)
		level.Monsters[to] = m
		m.Pos = to
	} else if to == level.Player.Pos{
		level.Attack(&m.Character, &level.Player.Character)
		fmt.Println("Monster attacked player")
		fmt.Println(m.Hitpoints, level.Player.Hitpoints)
		if m.Hitpoints <= 0 {
			delete(level.Monsters, m.Pos)
		}
		if level.Player.Hitpoints <= 0 {
			panic("YOU DIED")
		}
	}
}
