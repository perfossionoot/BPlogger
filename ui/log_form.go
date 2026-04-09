package ui

import (
	"image/color"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"BPlogger/bp"
	"BPlogger/db"
)

// classifiedField wraps an entry with a colored border rectangle and a label
// showing the current BP classification. Both update via OnChanged.
func classifiedField(entry *widget.Entry, classify func(int) bp.Classification) (
	field fyne.CanvasObject,
	borderRect *canvas.Rectangle,
	classLabel *widget.Label,
) {
	borderRect = canvas.NewRectangle(color.Transparent)
	// Stack: colored rect fills the full area, padded entry sits on top —
	// the 4 dp gap between them renders as a colored border.
	field = container.NewStack(borderRect, container.NewPadded(entry))
	classLabel = widget.NewLabel("")

	entry.OnChanged = func(s string) {
		val, _ := strconv.Atoi(s)
		cls := classify(val)
		borderRect.FillColor = cls.Color
		borderRect.Refresh()
		classLabel.SetText(cls.Label)
	}
	return
}

func NewLogForm(w fyne.Window, onSave func()) fyne.CanvasObject {
	systolicEntry := widget.NewEntry()
	systolicEntry.SetPlaceHolder("e.g. 120")
	sysBordered, _, sysLabel := classifiedField(systolicEntry, bp.ClassifySystolic)

	diastolicEntry := widget.NewEntry()
	diastolicEntry.SetPlaceHolder("e.g. 80")
	diaBordered, _, diaLabel := classifiedField(diastolicEntry, bp.ClassifyDiastolic)

	pulseEntry := widget.NewEntry()
	pulseEntry.SetPlaceHolder("e.g. 72")

	doSave := func() {
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
		}
		if err := db.InsertReading(r); err != nil {
			dialog.ShowError(err, w)
			return
		}

		systolicEntry.SetText("")
		diastolicEntry.SetText("")
		pulseEntry.SetText("")
		w.Canvas().Focus(systolicEntry)

		if onSave != nil {
			onSave()
		}
	}

	saveBtn := widget.NewButton("Save Reading", doSave)
	saveBtn.Importance = widget.HighImportance

	systolicEntry.OnSubmitted = func(_ string) { w.Canvas().Focus(diastolicEntry) }
	diastolicEntry.OnSubmitted = func(_ string) { w.Canvas().Focus(pulseEntry) }
	pulseEntry.OnSubmitted = func(_ string) { doSave() }

	form := widget.NewForm(
		widget.NewFormItem("Systolic (mmHg)", container.NewBorder(nil, nil, nil, sysLabel, sysBordered)),
		widget.NewFormItem("Diastolic (mmHg)", container.NewBorder(nil, nil, nil, diaLabel, diaBordered)),
		widget.NewFormItem("Pulse (bpm)", pulseEntry),
	)

	return container.NewVBox(form, saveBtn)
}
