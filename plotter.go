package main

import (
    "encoding/csv"
    "fmt"
    "os"
    "strconv"
    "time"
    "gonum.org/v1/plot"
    "gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/plotutil"
    "gonum.org/v1/plot/vg"
    "gonum.org/v1/plot/vg/draw" // Import the draw package
)

func main() {
    if err := plotWorkload("output.csv", "workload_plot.pdf"); err != nil {
        panic(err)
    }
}

func plotWorkload(csvFile, pdfFile string) error {
    f, err := os.Open(csvFile)
    if err != nil {
        return err
    }
    defer f.Close()

    r := csv.NewReader(f)
    records, err := r.ReadAll()
    if err != nil {
        return err
    }

    p := plot.New()
    p.Title.Text = "Workload Over Time"
    p.X.Label.Text = "Time"
    p.Y.Label.Text = "Workload"

    for i := 1; i < len(records[0]); i++ {
        var pts plotter.XYs
        for j, record := range records {
            if j == 0 {
                continue
            }
            t, err := time.Parse(time.RFC3339, record[0])
            if err != nil {
                return err
            }
            x := float64(t.Unix())
            y, err := strconv.ParseFloat(record[i], 64)
            if err != nil {
                return err
            }
            pts = append(pts, plotter.XY{X: x, Y: y})
        }
        line, points, err := plotter.NewLinePoints(pts)
        if err != nil {
            return err
        }
        line.Color = plotutil.Color(i)
        points.Shape = draw.CircleGlyph{} // Use draw.CircleGlyph
        p.Add(line, points)
        p.Legend.Add(fmt.Sprintf("Load %d", i), line, points)
    }

    p.X.Tick.Marker = plot.TimeTicks{Format: "15:04"}
    p.X.Tick.Length = vg.Points(10)
    p.Legend.Top = false
