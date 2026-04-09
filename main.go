package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"BPlogger/db"
	"BPlogger/ui"
)

func main() {
	if err := db.Init(); err != nil {
		log.Fatalf("failed to initialise database: %v", err)
	}
	defer db.DB.Close()

	a := app.New()
	w := a.NewWindow("BPlogger")
	w.Resize(fyne.NewSize(640, 600))

	list, refresh := ui.NewReadingsList(w)
	form := ui.NewLogForm(w, refresh)

	tabs := container.NewAppTabs(
		container.NewTabItem("Log Reading", form),
		container.NewTabItem("History", list),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	w.SetContent(tabs)
	w.ShowAndRun()
}
