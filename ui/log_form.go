package ui

import (
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"BPlogger/db"
)

func NewLogForm(w fyne.Window, onSave func()) fyne.CanvasObject {
	systolicEntry := widget.NewEntry()
	systolicEntry.SetPlaceHolder("e.g. 120")

	diastolicEntry := widget.NewEntry()
	diastolicEntry.SetPlaceHolder("e.g. 80")

	pulseEntry := widget.NewEntry()
	pulseEntry.SetPlaceHolder("e.g. 72")

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Optional notes...")
	notesEntry.SetMinRowsVisible(3)

	saveBtn := widget.NewButton("Save Reading", func() {
		sys, err1 := strconv.Atoi(systolicEntry.Text)
		dia, err2 := strconv.Atoi(diastolicEntry.Text)
		pul, err3 := strconv.Atoi(pulseEntry.Text)

		if err1 != nil || err2 != nil || err3 != nil || sys <= 0 || dia <= 0 || pul <= 0 {
			dialog.ShowError(errInvalidInput, w)
			return
		}

		r := db.Reading{
			Systolic:   sys,
			Diastolic:  dia,
			Pulse:      pul,
			RecordedAt: time.Now(),
			Notes:      notesEntry.Text,
		}
		if err := db.InsertReading(r); err != nil {
			dialog.ShowError(err, w)
			return
		}

		systolicEntry.SetText("")
		diastolicEntry.SetText("")
		pulseEntry.SetText("")
		notesEntry.SetText("")

		if onSave != nil {
			onSave()
		}
	})
	saveBtn.Importance = widget.HighImportance

	form := widget.NewForm(
		widget.NewFormItem("Systolic (mmHg)", systolicEntry),
		widget.NewFormItem("Diastolic (mmHg)", diastolicEntry),
		widget.NewFormItem("Pulse (bpm)", pulseEntry),
		widget.NewFormItem("Notes", notesEntry),
	)

	return container.NewVBox(form, saveBtn)
}
