package ui

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Chooser struct {
	*tview.Grid
	done  func(string, bool)
	dir   string
	files []string
}

func NewChooser(dir string, name string, done func(string, bool)) *Chooser {
	c := &Chooser{
		Grid: tview.NewGrid(),
		done: done,
		dir:  dir,
	}
	files, index := listGameFiles(dir, name)
	c.files = files
	list := tview.NewList().
		ShowSecondaryText(false).
		SetSelectedFunc(c.selectFile)
	list.SetBorder(true).SetTitle("Choose a game (ESC closes)")
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			c.done("", false)
			return nil
		}
		return event
	})
	for _, file := range files {
		list.AddItem(file, "", 0, nil)
	}
	list.AddItem("<new game>", "", 0, nil)
	if index >= 0 {
		list.SetCurrentItem(index)
	}
	c.SetColumns(2, 0, 2).SetRows(2, 0, 2).AddItem(list, 1, 1, 1, 1, 0, 0, true)
	return c
}

func (c *Chooser) selectFile(i int, _ string, _ string, _ rune) {
	switch {
	case i < len(c.files):
		c.done(path.Join(c.dir, c.files[i]), false)
	case len(c.files) == 0:
		c.done(path.Join(c.dir, fmt.Sprintf("%s-1.gm", time.Now().Format("20060102"))), false)
	default:
		c.done(path.Join(c.dir, c.files[len(c.files)-1]), true)
	}
}

func listGameFiles(dir string, name string) (result []string, index int) {
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".gm") {
			result = append(result, entry.Name())
		}
	}
	sort.Strings(result)
	index = -1
	for i := range result {
		if result[i] == name {
			index = i
			break
		}
	}
	return
}
