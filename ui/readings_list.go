package ui

import (
	"errors"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"BPlogger/db"
)

var errInvalidInput = errors.New("please enter valid positive numbers for systolic, diastolic, and pulse")

func NewReadingsList(w fyne.Window) (fyne.CanvasObject, func()) {
	var readings []db.Reading

	list := widget.NewList(
		func() int { return len(readings) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			r := readings[id]
			row := obj.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(
				fmt.Sprintf("%s  —  %d/%d mmHg  ❤ %d bpm",
					r.RecordedAt.Format("2006-01-02 15:04"),
					r.Systolic, r.Diastolic, r.Pulse,
				),
			)
			sub := row.Objects[1].(*widget.Label)
			var parts []string
			if len(r.Tags) > 0 {
				parts = append(parts, strings.Join(r.Tags, ", "))
			}
			if r.Notes != "" {
				parts = append(parts, r.Notes)
			}
			sub.SetText(strings.Join(parts, "  |  "))
			sub.Truncation = fyne.TextTruncateEllipsis
		},
	)

	var refresh func()

	list.OnSelected = func(id widget.ListItemID) {
		r := readings[id]
		tagsStr := "—"
		if len(r.Tags) > 0 {
			tagsStr = strings.Join(r.Tags, ", ")
		}
		msg := fmt.Sprintf(
			"Date:      %s\nSystolic:  %d mmHg\nDiastolic: %d mmHg\nPulse:     %d bpm\nTags:      %s\nNotes:     %s",
			r.RecordedAt.Format("2006-01-02 15:04:05"),
			r.Systolic, r.Diastolic, r.Pulse, tagsStr, r.Notes,
		)
		dialog.ShowConfirm("Reading Detail", msg+"\n\nDelete this reading?",
			func(del bool) {
				if del {
					if err := db.DeleteReading(r.ID); err != nil {
						dialog.ShowError(err, w)
					}
					refresh()
				}
				list.Unselect(id)
			}, w)
	}

	refresh = func() {
		var err error
		readings, err = db.GetReadings()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		list.Refresh()
	}

	refresh()
	return list, refresh
}
