package ui

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/hashicorp/go-multierror"
	"github.com/rivo/tview"
	"github.com/slshen/paperscore/pkg/boxscore"
	"github.com/slshen/paperscore/pkg/game"
	"github.com/slshen/paperscore/pkg/gamefile"
	"github.com/slshen/paperscore/pkg/stats"
)

type UI struct {
	Logger            *log.Logger
	RE                stats.RunExpectancy
	path              string
	app               *tview.Application
	root              *tview.Pages
	box               *tview.TextView
	properties        *LinedTextArea
	visitorPlaysStart int
	visitorPlays      *LinedTextArea
	homePlaysStart    int
	homePlays         *LinedTextArea
	messages          *tview.TextView
	lastKey           time.Time
	lastUpdate        time.Time
	focusOrder        []tview.Primitive
	status            *tview.TextView
	dialog            tview.Primitive
	modified          bool
}

var errorColor = tcell.ColorBlue

func New() *UI {
	ui := &UI{
		Logger:       log.New(io.Discard, "", 0),
		app:          tview.NewApplication(),
		root:         tview.NewPages(),
		properties:   NewLinedtextArea(),
		visitorPlays: NewLinedtextArea(),
		homePlays:    NewLinedtextArea(),
		box:          tview.NewTextView(),
		messages:     tview.NewTextView().SetDynamicColors(true),
		status:       tview.NewTextView().SetTextAlign(tview.AlignRight),
	}
	for _, box := range []any{ui.properties, ui.box, ui.visitorPlays, ui.homePlays, ui.messages} {
		box.(interface{ SetBorder(bool) *tview.Box }).SetBorder(true)
	}
	ui.focusOrder = []tview.Primitive{ui.properties, ui.visitorPlays, ui.homePlays}
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tview.NewFlex().
		AddItem(ui.properties, 0, 1, true).
		AddItem(ui.box, 0, 1, false),
		7, 0, true)
	flex.AddItem(tview.NewFlex().
		AddItem(ui.visitorPlays, 0, 1, false).
		AddItem(ui.homePlays, 0, 1, false),
		0, 1, false).
		AddItem(ui.messages, 6, 0, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewTextView().SetText("Quit:^Q Save:^S Choose:^L Swap:^R Box:^Z"), 0, 4, false).
			AddItem(ui.status, 0, 1, false),
			1, 0, false)
	ui.root.AddAndSwitchToPage("main", flex, true)
	ui.app.
		EnableMouse(true).
		SetRoot(ui.root, true).
		SetInputCapture(ui.inputHandler)
	return ui
}

func (ui *UI) update() {
	t0 := time.Now()
	var text string
	ui.app.QueueUpdate(func() {
		text = ui.getGameText()
	})
	file, err := gamefile.ParseString(ui.path, text)
	var (
		msg      string
		boxScore string
		gm       *game.Game
	)
	if err != nil {
		msg = err.Error()
	} else {
		gm, err = game.NewGame(file)
		if err != nil {
			msg = err.Error()
		} else {
			box, err := boxscore.NewBoxScore(gm, ui.RE)
			if err != nil {
				msg = err.Error()
			} else if box.HomeLineup.TeamStats != nil && box.VisitorLineup.TeamStats != nil {
				scoreTable := box.InningScoreTable()
				scoreTable.Columns[0].Name = ""
				rc := box.AltPlays().GetColumn("RCost").GetSummary()
				boxScore = fmt.Sprintf("%s\nMisplay runs = %.2f", scoreTable, rc)
			}
		}
	}
	ui.Logger.Println("updating message: ", msg)
	ui.Logger.Println("update took ", time.Since(t0))
	ui.app.QueueUpdateDraw(func() {
		ui.messages.SetText(msg)
		if boxScore != "" {
			ui.box.SetText(boxScore)
		}
		if gm != nil {
			ui.homePlays.SetTitle(gm.Home.Name)
			ui.visitorPlays.SetTitle(gm.Visitor.Name)
		}
		ui.properties.ClearColors()
		ui.homePlays.ClearColors()
		ui.visitorPlays.ClearColors()
		for _, err := range allErrors(err) {
			if gerr, ok := err.(interface{ Position() gamefile.Position }); ok {
				line := gerr.Position().Line - 1
				switch {
				case line < ui.visitorPlaysStart:
					ui.Logger.Println("highlighting properties line ", line)
					ui.properties.LineColors[line] = &errorColor
				case line < ui.homePlaysStart:
					ui.Logger.Println("highlighting visitor plays line ", line-ui.visitorPlaysStart)
					ui.visitorPlays.LineColors[line-ui.visitorPlaysStart] = &errorColor
				default:
					ui.Logger.Println("highlighting home plays line ", line-ui.homePlaysStart)
					ui.homePlays.LineColors[line-ui.homePlaysStart] = &errorColor
				}
			}
		}
	})
}

func allErrors(err error) []error {
	if err == nil {
		return nil
	}
	if m, ok := err.(*multierror.Error); ok && m.Len() > 0 {
		return m.Errors
	}
	return []error{err}
}

func (ui *UI) getGameText() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, ui.properties.GetText())
	if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != '\n' {
		fmt.Fprintln(&buf)
	}
	fmt.Fprintln(&buf, "---")
	fmt.Fprintln(&buf, "visitorplays")
	ui.visitorPlaysStart = lineCount(buf.String())
	visitorPlaysText := ui.visitorPlays.GetText()
	fmt.Fprint(&buf, visitorPlaysText)
	ui.homePlaysStart = ui.visitorPlaysStart + 1 + lineCount(visitorPlaysText)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		fmt.Fprintln(&buf)
		ui.homePlaysStart++
	}
	fmt.Fprintln(&buf, "homeplays")
	fmt.Fprintln(&buf, ui.homePlays.GetText())
	ui.lastUpdate = time.Now()
	ui.Logger.Println("got game text at ", ui.lastUpdate)
	return buf.String()
}

func lineCount(s string) (count int) {
	for _, ch := range s {
		if ch == '\n' {
			count++
		}
	}
	return count
}

func (ui *UI) parseGame(gamePath string) {
	ui.path = gamePath
	ui.status.SetText(path.Base(ui.path))
	ui.box.SetText("")
	ui.messages.SetText("")
	var r io.Reader
	f, err := os.Open(ui.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			r = strings.NewReader(fmt.Sprintf("date: %s\ngame: 1\n---\n", time.Now().Format(gamefile.GameDateFormat)))
		} else {
			ui.messages.SetText(fmt.Sprintf("cannot open %s: %s", ui.path, err))
			ui.properties.SetText("", false)
			ui.homePlays.SetText("", false)
			ui.visitorPlays.SetText("", false)
			return
		}
	} else {
		defer f.Close()
		r = f
	}
	scanner := bufio.NewScanner(r)
	state := "props"
	var (
		buf         strings.Builder
		targetPlays *LinedTextArea
	)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case state == "props":
			if line == "---" {
				state = "plays"
				ui.properties.Replace(0, ui.properties.GetTextLength(), buf.String())
				buf.Reset()
			} else {
				fmt.Fprintln(&buf, line)
			}
		case state == "plays" && line == "homeplays":
			state = "homeplays"
			targetPlays = ui.homePlays
		case state == "plays" && line == "visitorplays":
			state = "visitorplays"
			targetPlays = ui.visitorPlays
		case targetPlays != nil:
			if (line == "homeplays" || line == "visitorplays") && line != state {
				targetPlays.Replace(0, targetPlays.GetTextLength(), buf.String())
				buf.Reset()
				if state == "homeplays" {
					targetPlays = ui.visitorPlays
					state = "visitorplays"
				} else {
					targetPlays = ui.homePlays
					state = "homeplays"
				}
			} else {
				fmt.Fprintln(&buf, line)
			}
		}
	}
	if targetPlays != nil {
		targetPlays.Replace(0, targetPlays.GetTextLength(), buf.String())
	}
	// fake a key press so update cycle runs
	ui.lastKey = time.Now()
	ui.modified = false
}

func (ui *UI) chooseFile() {
	ui.showDialog(NewChooser(path.Dir(ui.path), ui.path, func(s string, newgame bool) {
		ui.closeDialog()
		if s != "" {
			if newgame {
				ui.newGame(s)
			} else {
				ui.parseGame(s)
			}
		}
	}))
}

func (ui *UI) swapHomeAndAway() {
	var propertiesText strings.Builder
	for _, line := range strings.Split(ui.properties.GetText(), "\n") {
		if strings.HasPrefix(line, "home") && len(line) > 4 {
			line = "visitor" + line[4:]
		} else if strings.HasPrefix(line, "visitor") && len(line) > 7 {
			line = "home" + line[7:]
		}
		fmt.Fprintln(&propertiesText, line)
	}
	ui.properties.SetText(propertiesText.String(), false)
	visitorPlays, homePlays := ui.visitorPlays.GetText(), ui.homePlays.GetText()
	ui.visitorPlays.Replace(0, ui.visitorPlays.GetTextLength(), homePlays)
	ui.homePlays.Replace(0, ui.homePlays.GetTextLength(), visitorPlays)
	ui.lastKey = time.Now()
}

func (ui *UI) inputHandler(event *tcell.EventKey) *tcell.EventKey {
	if ui.dialog != nil {
		return event
	}
	var focusInc int
	switch event.Key() {
	case tcell.KeyCtrlQ:
		ui.app.Stop()
	case tcell.KeyCtrlC:
		return nil
	case tcell.KeyCtrlS:
		if ui.dialog == nil {
			ui.save()
		}
		return nil
	case tcell.KeyCtrlL:
		if ui.dialog == nil {
			ui.chooseFile()
		}
		return nil
	case tcell.KeyCtrlR:
		ui.swapHomeAndAway()
	case tcell.KeyCtrlZ:
		ui.showBox()
		return nil
	case tcell.KeyCtrlP:
		return tcell.NewEventKey(tcell.KeyUp, 0, 0)
	case tcell.KeyCtrlN:
		return tcell.NewEventKey(tcell.KeyDown, 0, 0)
	case tcell.KeyCtrlF:
		return tcell.NewEventKey(tcell.KeyRight, 0, 0)
	case tcell.KeyCtrlB:
		return tcell.NewEventKey(tcell.KeyLeft, 0, 0)
	case tcell.KeyUp:
	case tcell.KeyDown:
	case tcell.KeyLeft:
	case tcell.KeyRight:
		break
	case tcell.KeyTAB:
		focusInc = 1
	case tcell.KeyBacktab:
		focusInc = -1
	default:
		switch {
		case ui.properties.HasFocus():
			fallthrough
		case ui.homePlays.HasFocus():
			fallthrough
		case ui.visitorPlays.HasFocus():
			if !ui.modified {
				ui.modified = true
				ui.status.SetText(ui.status.GetText(false) + "*")
			}
			ui.lastKey = event.When()
			ui.Logger.Println("got key at ", ui.lastKey)
		}
	}
	if focusInc != 0 {
		j := 0
		for i := range ui.focusOrder {
			if ui.focusOrder[i].HasFocus() {
				j = i + focusInc
				if j < 0 {
					j = len(ui.focusOrder) - 1
				} else if j == len(ui.focusOrder) {
					j = 0
				}
			}
		}
		ui.app.SetFocus(ui.focusOrder[j])
		return nil
	}
	return event
}

func (ui *UI) newGame(gamePath string) {
	file, _ := gamefile.ParseFile(gamePath)
	if file != nil {
		newFile, _ := file.WriteNewGame(false)
		if newFile != nil {
			ui.parseGame(newFile.Path)
			return
		}
	}
	m := regexp.MustCompile(`(^.*-)([0-9]+)\.gm$`).FindStringSubmatch(gamePath)
	var newGamePath string
	if m != nil {
		n, _ := strconv.Atoi(m[2])
		if n != 0 {
			newGamePath = m[1] + fmt.Sprintf("%d", n+1) + ".gm"
		}
	}
	if newGamePath == "" {
		newGamePath = path.Join(path.Dir(gamePath), time.Now().Format(gamefile.GameDateFormat)+"-1.gm")
	}
	ui.homePlays.SetText("", false)
	ui.visitorPlays.SetText("", false)
	ui.parseGame(newGamePath)
}

func (ui *UI) save() {
	text := ui.getGameText()
	var canonName string
	if gf, err := gamefile.ParseString(ui.path, text); err == nil {
		gm, _ := game.NewGame(gf)
		if gm != nil && gm.GetDate().Unix() != 0 && gm.Number != "" {
			canonName = fmt.Sprintf("%s-%s.gm", gm.GetDate().Format("20060102"), gm.Number)
			if canonName == path.Base(ui.path) {
				canonName = ""
			}
		}
		var s strings.Builder
		gf.Write(&s)
		text = s.String()
	}
	doSave := func() {
		originalPath := ui.path
		if canonName != "" {
			ui.path = path.Join(path.Dir(ui.path), canonName)
		}
		f, err := os.CreateTemp(path.Dir(ui.path), fmt.Sprintf("%s*", path.Base(ui.path)))
		if err != nil {
			msg := fmt.Sprintf("cannot save %s [yellow:red]%s", ui.path, err.Error())
			ui.messages.SetText(msg)
		} else {
			_, _ = f.WriteString(text)
			f.Close()
			if err := os.Rename(f.Name(), ui.path); err != nil {
				ui.messages.SetText(fmt.Sprintf("could not save %s [yellow:red]%s", ui.path, err.Error()))
			} else {
				ui.parseGame(ui.path)
				if canonName != "" {
					_ = os.Remove(originalPath)
				}
			}
		}
	}
	if canonName != "" {
		ui.showQuestionDialog(fmt.Sprintf("Rename %s to %s ?", path.Base(ui.path), canonName), "OK", doSave)
	} else {
		doSave()
	}
}

func (ui *UI) showDialog(dialog tview.Primitive) {
	if ui.dialog == nil {
		ui.dialog = dialog
		ui.root.AddPage("dialog", dialog, true, true)
		ui.app.SetFocus(dialog)
	}
}

func (ui *UI) closeDialog() {
	ui.root.RemovePage("dialog")
	ui.dialog = nil
}

func (ui *UI) showQuestionDialog(question string, okLabel string, ok func()) {
	modal := tview.NewModal().AddButtons([]string{okLabel, "Cancel"}).
		SetText(question).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == okLabel {
				ok()
			}
			ui.closeDialog()
		})
	ui.showDialog(modal)
}

func (ui *UI) showBox() {
	gf, err := gamefile.ParseString(ui.path, ui.getGameText())
	if err != nil {
		ui.messages.SetText(fmt.Sprintf("gamefile is invalid: %s", err.Error()))
		return
	}
	g, err := game.NewGame(gf)
	if g == nil {
		ui.messages.SetText(fmt.Sprintf("game is invalid: %s", err.Error()))
		return
	}
	box, err := boxscore.NewBoxScore(g, ui.RE)
	if err != nil {
		ui.messages.SetText(fmt.Sprintf("could not generate boxscore: %s", err.Error()))
		return
	}
	box.IncludePlays = true
	var s strings.Builder
	_ = box.Render(&s)
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlQ {
			ui.closeDialog()
			return nil
		}
		return event
	})
	view := tview.NewTextView().SetText(s.String())
	view.SetBorder(true)
	flex.AddItem(view, 0, 1, true).
		AddItem(tview.NewTextView().SetText("Close:^Q"), 1, 0, false)
	ui.showDialog(flex)
}

func (ui *UI) backgroundUpdate(ticker *time.Ticker, done <-chan bool) {
	for {
		select {
		case <-ticker.C:
			ui.app.QueueUpdate(func() {
				if ui.lastKey.After(ui.lastUpdate) {
					go ui.update()
				}
			})
		case <-done:
			return
		}
	}
}

func (ui *UI) Run(path string) error {
	ticker := time.NewTicker(250 * time.Millisecond)
	done := make(chan bool)
	go ui.backgroundUpdate(ticker, done)
	if path == "" {
		path = "."
	}
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		ui.showDialog(NewChooser(path, "", func(s string, newgame bool) {
			if s == "" {
				ui.app.Stop()
			} else {
				ui.closeDialog()
				if newgame {
					ui.newGame(s)
				} else {
					ui.parseGame(s)
				}
			}
		}))
	} else {
		ui.parseGame(path)
	}
	err = ui.app.Run()
	ticker.Stop()
	done <- true
	return err
}
