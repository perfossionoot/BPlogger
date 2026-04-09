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
	w.Resize(fyne.NewSize(640, 640))

	trendsContent, updateTrends := ui.NewTrendsTab()

	reloadTrends := func() {
		readings, _ := db.GetReadings()
		updateTrends(readings)
	}

	list, listRefresh := ui.NewReadingsList(w)

	onSave := func() {
		listRefresh()
		reloadTrends()
	}

	form := ui.NewLogForm(w, onSave)

	reloadTrends() // initial chart load

	tabs := container.NewAppTabs(
		container.NewTabItem("Log Reading", form),
		container.NewTabItem("History", list),
		container.NewTabItem("Trends", trendsContent),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	w.SetContent(tabs)
	w.ShowAndRun()
}
