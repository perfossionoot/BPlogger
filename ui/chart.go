package ui

import (
	"fmt"
	"image"
	"image/color"
	"sort"

	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"BPlogger/db"
)

const (
	chartYMin = 50
	chartYMax = 220
)

var chartZones = []struct {
	lo, hi int
	col    color.RGBA
}{
	{chartYMin, 120, color.RGBA{76, 175, 80, 40}},
	{120, 130, color.RGBA{255, 193, 7, 40}},
	{130, 140, color.RGBA{255, 152, 0, 40}},
	{140, 180, color.RGBA{220, 53, 69, 40}},
	{180, chartYMax, color.RGBA{136, 0, 0, 40}},
}

// BPChart renders a BP trend chart as a Fyne raster widget.
type BPChart struct {
	widget.BaseWidget
	readings []db.Reading
}

func NewBPChart() *BPChart {
	c := &BPChart{}
	c.ExtendBaseWidget(c)
	return c
}

func (c *BPChart) SetReadings(readings []db.Reading) {
	sorted := make([]db.Reading, len(readings))
	copy(sorted, readings)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RecordedAt.Before(sorted[j].RecordedAt)
	})
	c.readings = sorted
	c.Refresh()
}

func (c *BPChart) MinSize() fyne.Size { return fyne.NewSize(400, 280) }

func (c *BPChart) CreateRenderer() fyne.WidgetRenderer {
	raster := canvas.NewRaster(func(w, h int) image.Image {
		return paintBPChart(c.readings, w, h)
	})
	return widget.NewSimpleRenderer(raster)
}

// NewTrendsTab returns the Trends tab content and an update func.
func NewTrendsTab() (fyne.CanvasObject, func([]db.Reading)) {
	chart := NewBPChart()
	stats := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{})

	update := func(readings []db.Reading) {
		chart.SetReadings(readings)
		if len(readings) == 0 {
			stats.SetText("")
			return
		}
		var sumSys, sumDia int
		earliest, latest := readings[0].RecordedAt, readings[0].RecordedAt
		for _, r := range readings {
			sumSys += r.Systolic
			sumDia += r.Diastolic
			if r.RecordedAt.Before(earliest) {
				earliest = r.RecordedAt
			}
			if r.RecordedAt.After(latest) {
				latest = r.RecordedAt
			}
		}
		n := len(readings)
		stats.SetText(fmt.Sprintf(
			"Average %d/%d mmHg  •  %d readings  •  %s – %s",
			sumSys/n, sumDia/n, n,
			earliest.Format("Jan 2"),
			latest.Format("Jan 2, 2006"),
		))
	}

	return container.NewBorder(nil, stats, nil, nil, chart), update
}

// ── paint ────────────────────────────────────────────────────────────────────

func paintBPChart(readings []db.Reading, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Estimate HiDPI scale from physical pixel width vs logical MinSize (400).
	scale := min(2, max(1, w/400))

	lPad := 44 * scale
	rPad := 12 * scale
	tPad := 20 * scale
	bPad := 32 * scale
	charH := 13 * scale

	solidFill(img, 0, 0, w, h, color.RGBA{248, 248, 248, 255})

	px0, px1 := lPad, w-rPad
	py0, py1 := tPad, h-bPad
	cw, ch := px1-px0, py1-py0
	if cw <= 0 || ch <= 0 {
		return img
	}

	toY := func(mm int) int {
		ratio := float64(chartYMax-mm) / float64(chartYMax-chartYMin)
		return py0 + int(ratio*float64(ch))
	}

	// Zone bands
	for _, z := range chartZones {
		blendFill(img, px0, toY(z.hi), px1, toY(z.lo), z.col)
	}

	// Grid lines + Y-axis labels
	gridCol := color.RGBA{160, 160, 160, 120}
	lblCol := color.RGBA{80, 80, 80, 255}
	for _, mm := range []int{80, 90, 120, 130, 140, 180} {
		y := toY(mm)
		hline(img, px0, px1, y, gridCol)
		drawScaledText(img, fmt.Sprintf("%d", mm), 2*scale, y-charH/2, scale, lblCol)
	}

	// Axes
	axisCol := color.RGBA{100, 100, 100, 255}
	vline(img, px0, py0, py1, axisCol)
	hline(img, px0, px1, py1, axisCol)

	sysCol := color.RGBA{180, 40, 40, 230}
	diaCol := color.RGBA{30, 100, 200, 230}

	// Legend
	ly := tPad/2 + charH/2
	for dy := 0; dy < max(1, scale); dy++ {
		hline(img, px0+4*scale, px0+16*scale, ly+dy, sysCol)
		hline(img, px0+80*scale, px0+92*scale, ly+dy, diaCol)
	}
	drawScaledText(img, "Systolic", px0+18*scale, ly-charH/2, scale, sysCol)
	drawScaledText(img, "Diastolic", px0+94*scale, ly-charH/2, scale, diaCol)

	if len(readings) == 0 {
		drawScaledText(img, "No readings yet", px0+cw/4, py0+ch/2, scale, color.RGBA{150, 150, 150, 255})
		return img
	}

	n := len(readings)
	xAt := func(i int) int {
		if n == 1 {
			return px0 + cw/2
		}
		return px0 + i*(cw-1)/(n-1)
	}

	// Lines
	for i := 1; i < n; i++ {
		thickLine(img, xAt(i-1), toY(readings[i-1].Systolic), xAt(i), toY(readings[i].Systolic), sysCol, scale)
		thickLine(img, xAt(i-1), toY(readings[i-1].Diastolic), xAt(i), toY(readings[i].Diastolic), diaCol, scale)
	}

	// Dots
	dotR := 4 * scale
	for i, r := range readings {
		fillCircle(img, xAt(i), toY(r.Systolic), dotR, sysCol)
		fillCircle(img, xAt(i), toY(r.Diastolic), dotR, diaCol)
	}

	// X-axis date labels (at most 8)
	step := max(1, n/8)
	for i := 0; i < n; i += step {
		lbl := readings[i].RecordedAt.Format("01/02")
		cx := xAt(i) - 7*scale*len(lbl)/2
		drawScaledText(img, lbl, cx, py1+charH/2, scale, lblCol)
	}

	return img
}

// ── drawing helpers ───────────────────────────────────────────────────────────

func solidFill(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	b := img.Bounds()
	for y := max(y0, b.Min.Y); y < min(y1, b.Max.Y); y++ {
		for x := max(x0, b.Min.X); x < min(x1, b.Max.X); x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func blendFill(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	b := img.Bounds()
	a := float64(c.A) / 255.0
	for y := max(y0, b.Min.Y); y < min(y1, b.Max.Y); y++ {
		for x := max(x0, b.Min.X); x < min(x1, b.Max.X); x++ {
			d := img.RGBAAt(x, y)
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(float64(c.R)*a + float64(d.R)*(1-a)),
				G: uint8(float64(c.G)*a + float64(d.G)*(1-a)),
				B: uint8(float64(c.B)*a + float64(d.B)*(1-a)),
				A: 255,
			})
		}
	}
}

func hline(img *image.RGBA, x0, x1, y int, c color.RGBA) {
	b := img.Bounds()
	if y < b.Min.Y || y >= b.Max.Y {
		return
	}
	for x := max(x0, b.Min.X); x < min(x1, b.Max.X); x++ {
		img.SetRGBA(x, y, c)
	}
}

func vline(img *image.RGBA, x, y0, y1 int, c color.RGBA) {
	b := img.Bounds()
	if x < b.Min.X || x >= b.Max.X {
		return
	}
	for y := max(y0, b.Min.Y); y < min(y1, b.Max.Y); y++ {
		img.SetRGBA(x, y, c)
	}
}

func bresenham(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx, dy := abs(x1-x0), abs(y1-y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	b := img.Bounds()
	for {
		if x0 >= b.Min.X && x0 < b.Max.X && y0 >= b.Min.Y && y0 < b.Max.Y {
			img.SetRGBA(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func thickLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA, scale int) {
	thickness := max(1, scale)
	for t := 0; t < thickness; t++ {
		bresenham(img, x0, y0+t, x1, y1+t, c)
	}
}

func fillCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	b := img.Bounds()
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			if (x-cx)*(x-cx)+(y-cy)*(y-cy) <= r*r {
				if x >= b.Min.X && x < b.Max.X && y >= b.Min.Y && y < b.Max.Y {
					img.SetRGBA(x, y, c)
				}
			}
		}
	}
}

// drawScaledText draws text at 1x and scales up pixel-by-pixel for HiDPI.
func drawScaledText(img *image.RGBA, text string, x, y, scale int, c color.RGBA) {
	charW, charH := 7, 13
	tmpW := len(text)*charW + 2
	tmp := image.NewRGBA(image.Rect(0, 0, tmpW, charH))
	d := xfont.Drawer{
		Dst:  tmp,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(0, 11), // baseline at row 11
	}
	d.DrawString(text)

	b := img.Bounds()
	for py := 0; py < charH; py++ {
		for px := 0; px < tmpW; px++ {
			if tmp.RGBAAt(px, py).A == 0 {
				continue
			}
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					ix, iy := x+px*scale+sx, y+py*scale+sy
					if ix >= b.Min.X && ix < b.Max.X && iy >= b.Min.Y && iy < b.Max.Y {
						img.SetRGBA(ix, iy, c)
					}
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
