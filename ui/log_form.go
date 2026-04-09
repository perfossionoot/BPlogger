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

	availableTags := []string{"poor sleep", "stressed", "bp- meds"}
	tagChecks := make([]*widget.Check, len(availableTags))
	for i, t := range availableTags {
		tagChecks[i] = widget.NewCheck(t, nil)
	}
	tagsRow := container.NewHBox(tagChecks[0], tagChecks[1], tagChecks[2])

	systolicEntry.OnSubmitted = func(_ string) { w.Canvas().Focus(diastolicEntry) }
	diastolicEntry.OnSubmitted = func(_ string) { w.Canvas().Focus(pulseEntry) }
	pulseEntry.OnSubmitted = func(_ string) { w.Canvas().Focus(notesEntry) }

	saveBtn := widget.NewButton("Save Reading", func() {
		sys, err1 := strconv.Atoi(systolicEntry.Text)
		dia, err2 := strconv.Atoi(diastolicEntry.Text)
		pul, err3 := strconv.Atoi(pulseEntry.Text)

		if err1 != nil || err2 != nil || err3 != nil || sys <= 0 || dia <= 0 || pul <= 0 {
			dialog.ShowError(errInvalidInput, w)
			return
		}

		var tags []string
		for i, chk := range tagChecks {
			if chk.Checked {
				tags = append(tags, availableTags[i])
			}
		}

		r := db.Reading{
			Systolic:   sys,
			Diastolic:  dia,
			Pulse:      pul,
			RecordedAt: time.Now(),
			Tags:       tags,
			Notes:      notesEntry.Text,
		}
		if err := db.InsertReading(r); err != nil {
			dialog.ShowError(err, w)
			return
		}

		systolicEntry.SetText("")
		diastolicEntry.SetText("")
		pulseEntry.SetText("")
		for _, chk := range tagChecks {
			chk.SetChecked(false)
		}
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
		widget.NewFormItem("Tags", tagsRow),
		widget.NewFormItem("Notes", notesEntry),
	)

	return container.NewVBox(form, saveBtn)
}
