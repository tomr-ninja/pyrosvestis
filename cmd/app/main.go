package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"

	"github.com/tomr-ninja/pyrosvestis"
)

const endpoint = "http://10.4.8.17:89/metrics"

func main() {
	l := tview.NewList().SetSelectedFocusOnly(true)
	t := tview.NewTable().SetBorders(true)
	t.SetTitleAlign(tview.AlignLeft)

	grid := tview.NewGrid().
		SetRows(8, 16).
		SetBorders(true).
		AddItem(l, 0, 0, 1, 1, 0, 0, false).
		AddItem(t, 1, 0, 1, 1, 0, 0, false)

	// fetch metrics from the endpoint
	s, err := pyrosvestis.FetchMetrics(endpoint)
	if err != nil {
		panic(fmt.Sprintf("failed to fetch metrics: %s", err))
	}

	for _, metric := range s.Metrics {
		l.AddItem(metric.Name, "", 0, nil)
	}

	app := tview.NewApplication().
		SetRoot(grid, true).
		SetFocus(l)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyLeft:
			app.SetFocus(l)
			t.SetSelectable(false, false)
			t.Select(0, 0)
			t.Clear()

			return nil
		case tcell.KeyEnter, tcell.KeyRight:
			mg := s.Metrics[l.GetCurrentItem()]
			t.SetTitle(fmt.Sprintf("Metric Details (%s)", mg.Name))
			t.Clear()

			populateTable(mg, t)
			t.ScrollToBeginning()
			t.SetSelectable(true, false)

			app.SetFocus(t)

			return nil

		default:
			return event
		}
	})

	if err = app.Run(); err != nil {
		panic(err)
	}
}

func populateTable(src pyrosvestis.MetricsGroup, dst *tview.Table) {
	// first, determine the table shape (cols, rows)
	var (
		cols, rows int
		headers    = make([]string, 0, 10)
		index      = map[string]int{}
	)

	for _, _m := range src.Metrics {
		for _, m := range _m.Metric {
			for _, label := range m.Label {
				if _, ok := index[*label.Name]; !ok {
					index[*label.Name] = cols
					headers = append(headers, *label.Name)
					cols++
				}
			}
			rows++
		}
	}

	var (
		labels    = make([][]string, 0, rows)
		summaries = make([]string, 0, rows)
	)
	for _, mf := range src.Metrics {
		for _, m := range mf.Metric {
			ml := make([]string, 0, cols)
			for _, label := range m.Label {
				idx := index[*label.Name]

				// ensure the slice is large enough
				if len(ml) <= idx {
					ml = append(ml, make([]string, idx-len(ml)+1)...)
				}
				ml[idx] = *label.Value
			}
			labels = append(labels, ml)
			switch *mf.Type {
			case dto.MetricType_COUNTER:
				summaries = append(summaries, fmt.Sprintf("V: %d", int(m.Counter.GetValue())))
			case dto.MetricType_GAUGE:
				summaries = append(summaries, fmt.Sprintf("V: %d", int(m.Counter.GetValue())))
			case dto.MetricType_SUMMARY:
				summaries = append(summaries, "")
			case dto.MetricType_HISTOGRAM:
				summaries = append(summaries, "")
			}
		}
	}

	headers = append(headers, "Summary")
	for i := range headers {
		dst.SetCellSimple(0, i, headers[i])
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			dst.SetCellSimple(i+1, j, labels[i][j])
		}
		dst.SetCellSimple(i+1, cols, summaries[i])
	}
}
