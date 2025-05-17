package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
)

var screen tcell.Screen

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range []rune(text) {
		screen.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}

func drawBox(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// Fill background
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			screen.SetContent(col, row, ' ', nil, style)
		}
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		screen.SetContent(col, y1, tcell.RuneHLine, nil, style)
		screen.SetContent(col, y2, tcell.RuneHLine, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		screen.SetContent(x1, row, tcell.RuneVLine, nil, style)
		screen.SetContent(x2, row, tcell.RuneVLine, nil, style)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		screen.SetContent(x1, y1, tcell.RuneULCorner, nil, style)
		screen.SetContent(x2, y1, tcell.RuneURCorner, nil, style)
		screen.SetContent(x1, y2, tcell.RuneLLCorner, nil, style)
		screen.SetContent(x2, y2, tcell.RuneLRCorner, nil, style)
	}

	drawText(screen, x1+1, y1+1, x2-1, y2-1, style, text)
}

func main() {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	boxStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorPurple)
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	// Initialize screen
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	screen.SetStyle(defStyle)
	screen.EnableMouse()
	screen.EnablePaste()
	screen.Clear()

	// Draw initial boxes
	//drawBox(screen, 1, 1, 42, 7, boxStyle, "Click and drag to draw a box")
	//drawBox(screen, 5, 9, 32, 14, boxStyle, "Press C to reset")
	drawText(screen, 1, 1, 42, 3, textStyle, "Hey there")
	//drawText(screen, 1, 4, 42, 20, textStyle, "0 seconds")

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	// Here's how to get the screen size when you need it.
	// xmax, ymax := screen.Size()

	// Here's an example of how to inject a keystroke where it will
	// be picked up by the next PollEvent call.  Note that the
	// queue is LIFO, it has a limited length, and PostEvent() can
	// return an error.
	// screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, rune('a'), 0))

	go func() {
		i := 1
		for {
			drawText(screen, 1, 3, 42, 6, textStyle, fmt.Sprintf("%v seconds", i))
			// Update screen
			screen.Show()
			log.Println("iteration", i)
			time.Sleep(1 * time.Second)
			i += 1
		}
	}()
	// Event loop
	ox, oy := -1, -1
	for {
		// Update screen
		screen.Show()

		// Poll event
		ev := screen.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC || ev.Rune() == 'q' || ev.Rune() == 'Q' {
				return
			} else if ev.Key() == tcell.KeyCtrlL {
				screen.Sync()
			} else if ev.Rune() == 'C' || ev.Rune() == 'c' {
				screen.Clear()
			}
		case *tcell.EventMouse:
			x, y := ev.Position()

			switch ev.Buttons() {
			case tcell.Button1, tcell.Button2:
				if ox < 0 {
					ox, oy = x, y // record location when click started
				}

			case tcell.ButtonNone:
				if ox >= 0 {
					label := fmt.Sprintf("%d,%d to %d,%d", ox, oy, x, y)
					drawBox(screen, ox, oy, x, y, boxStyle, label)
					ox, oy = -1, -1
				}
			}
		}
	}
}
