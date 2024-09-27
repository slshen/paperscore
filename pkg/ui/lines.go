package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TextArea doesn't provide GetSelectedStyle() so just assume it's the default
var defaultSelectedStyle = tcell.StyleDefault.Background(tview.Styles.PrimaryTextColor).
	Foreground(tview.Styles.PrimitiveBackgroundColor)

type LinedTextArea struct {
	*tview.TextArea
	LineColors map[int]*tcell.Color
}

func NewLinedtextArea() *LinedTextArea {
	t := &LinedTextArea{
		TextArea:   tview.NewTextArea(),
		LineColors: map[int]*tcell.Color{},
	}
	return t
}

func (t *LinedTextArea) ClearColors() {
	for k := range t.LineColors {
		delete(t.LineColors, k)
	}
}

func (t *LinedTextArea) Draw(screen tcell.Screen) {
	t.TextArea.Draw(screen)
	firstLine, _ := t.GetOffset()
	x, y, width, height := t.GetInnerRect()
	for row := y; row < y+height; row++ {
		line := firstLine + row - y
		lineColor := t.LineColors[line]
		if lineColor != nil {
			for col := x; col < x+width; {
				ch, chc, chStyle, chw := screen.GetContent(col, row)
				if chStyle == defaultSelectedStyle {
					chStyle = chStyle.Foreground(*lineColor)
				} else {
					chStyle = chStyle.Background(*lineColor)
				}
				screen.SetContent(col, row, ch, chc, chStyle)
				col += chw
			}
		}
	}
}
