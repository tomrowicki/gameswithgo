package ui2d

import (
	"fmt"
	. "gameswithgo/rpg/game"
	"github.com/veandco/go-sdl2/sdl"
)

func (ui *ui) DrawInventory(level *Level) {
	playerSrcRect := ui.textureIndex[level.Player.Rune][0]
	invRect := ui.getInventoryRect()
	ui.renderer.Copy(ui.groundInventoryBackground, nil, invRect)
	offset := int32(float64(invRect.H) * .05)

	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect, &sdl.Rect{invRect.X + invRect.X/4, invRect.Y + offset, invRect.W / 2, invRect.H / 2})
	ui.renderer.Copy(ui.slotBackground, nil, ui.getHelmetSlotRect())
	if level.Player.Helmet != nil {
		ui.renderer.Copy(ui.textureAtlas, &ui.textureIndex[level.Player.Helmet.Rune][0], ui.getHelmetSlotRect())
	}
	ui.renderer.Copy(ui.slotBackground, nil, ui.getWeaponSlotRect())
	if level.Player.Weapon != nil {
		ui.renderer.Copy(ui.textureAtlas, &ui.textureIndex[level.Player.Weapon.Rune][0], ui.getWeaponSlotRect())
	}

	for i, item := range level.Player.Items {
		itemSrcRect := ui.textureIndex[item.Rune][0]
		if item == ui.draggedItem {
			itemSize := int32(float32(ui.winWidth) * itemSizeRatio)
			ui.renderer.Copy(ui.textureAtlas, &itemSrcRect, &sdl.Rect{int32(ui.currMouseState.pos.X), int32(ui.currMouseState.pos.Y), itemSize, itemSize})
		} else {
			ui.renderer.Copy(ui.textureAtlas, &itemSrcRect, ui.getInventoryItemRect(i))
		}
	}
}

func (ui *ui) getHelmetSlotRect() *sdl.Rect {
	invRect := ui.getInventoryRect()
	slotSize := int32(itemSizeRatio * float32(ui.winWidth) * 1.05)
	return &sdl.Rect{(invRect.X*2+invRect.W)/2 - slotSize/2, invRect.Y, slotSize, slotSize}
}

func (ui *ui) getWeaponSlotRect() *sdl.Rect {
	invRect := ui.getInventoryRect()
	slotSize := int32(itemSizeRatio * float32(ui.winWidth) * 1.05)
	yOffset := int32(float64(invRect.H) * .17)
	xOffset := int32(float64(invRect.W) * .17)
	return &sdl.Rect{invRect.X + xOffset, invRect.Y + yOffset, slotSize, slotSize}
}

func (ui *ui) getInventoryRect() *sdl.Rect {
	invWidth := int32(float32(ui.winWidth) * .40)
	invHeight := int32(float32(ui.winHeight) * .75)
	offsetX := (int32(ui.winWidth) - invWidth) / 2
	offsetY := (int32(ui.winHeight) - invHeight) / 2
	return &sdl.Rect{offsetX, offsetY, invWidth, invHeight}
}

func (ui *ui) getInventoryItemRect(i int) *sdl.Rect {
	invRect := ui.getInventoryRect()
	itemSize := int32(float32(ui.winWidth) * itemSizeRatio)
	return &sdl.Rect{invRect.X + int32(i)*itemSize, invRect.Y + invRect.H - itemSize, itemSize, itemSize}
}

func (ui *ui) CheckEquippedItem() *Item {
	if ui.draggedItem.Typ == Weapon {
		r := ui.getWeaponSlotRect()
		if r.HasIntersection(&sdl.Rect{int32(ui.currMouseState.pos.X), int32(ui.currMouseState.pos.Y), 1,1}) {
			return ui.draggedItem
		}
	} else if ui.draggedItem.Typ == Helmet {
		r := ui.getHelmetSlotRect()
		if r.HasIntersection(&sdl.Rect{int32(ui.currMouseState.pos.X), int32(ui.currMouseState.pos.Y), 1,1}) {
			return ui.draggedItem
		}
	}
	return nil
}

func (ui *ui) CheckDroppedItem() *Item {
	invRect := ui.getInventoryRect()
	mousePos := ui.currMouseState.pos
	if invRect.HasIntersection(&sdl.Rect{int32(mousePos.X), int32(mousePos.Y), 1, 1}) {
		return nil
	}
	return ui.draggedItem
}

func (ui *ui) CheckInventoryItems(level *Level) *Item {
	if ui.currMouseState.leftButton {
		mousePos := ui.currMouseState.pos
		for i, item := range level.Player.Items {
			itemRect := ui.getInventoryItemRect(i)
			// checking if click occurs within the item's rect
			if itemRect.HasIntersection(&sdl.Rect{int32(mousePos.X), int32(mousePos.Y), 1, 1}) {
				fmt.Println("clicked inventory item!")
				return item
			}
		}
	}
	return nil
}

func (ui *ui) CheckGroundItems(level *Level) *Item {
	if !ui.currMouseState.leftButton && ui.prevMouseState.leftButton {
		items := level.Items[level.Player.Pos]
		mousePos := ui.currMouseState.pos
		for i, item := range items {
			itemRect := ui.getGroundItemRect(i)
			// checking if click occurs within the item's rect
			if itemRect.HasIntersection(&sdl.Rect{int32(mousePos.X), int32(mousePos.Y), 1, 1}) {
				return item
			}
		}
	}
	return nil
}
