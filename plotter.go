package main

import (
    "bufio"
    "encoding/csv"
    "fmt"
    "os"
    "strconv"
    "strings"
    "image/color"
    "time"
    "gonum.org/v1/plot"
    "gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/plotutil"
    "gonum.org/v1/plot/vg"
    "gonum.org/v1/plot/vg/draw"
    "math/rand"
)
func init() {
    rand.Seed(time.Now().UnixNano())
}
func main() {
    reader := bufio.NewReader(os.Stdin)
    fmt.Println("Select plot type: 'all' for all workloads, 'individual' for individual workloads, 'vol_interval' for volatility intervals, 'changes' for workload changes")
    fmt.Print("Enter choice: ")
    choice, _ := reader.ReadString('\n')
    choice = strings.TrimSpace(choice) // Trim whitespace and newline character

    switch choice {
    case "all":
        if err := plotAllWorkloads("output.csv", "all_workloads_plot.pdf"); err != nil {
            panic(err)
        }
    case "individual":
        if err := plotWorkload("output.csv", "workload_plot.pdf"); err != nil {
            panic(err)
        }
    case "vol_interval":
        if err := plotWorkloadVolatilityIntervals("volatility_output.csv", "volatility_intervals_plot.pdf"); err != nil {
            panic(err)
        }
    case "changes":
        if err := plotWorkloadChanges("workload_volatility_intervals.csv", "workload_changes_plot.png"); err != nil {
            panic(err)
        }
    default:
        fmt.Println("Invalid choice. Please run the program again.")
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
    p.Legend.Left = false

    if err := p.Save(18*vg.Inch, 6*vg.Inch, pdfFile); err != nil {
        return err
    }
    return nil
}
func plotAllWorkloads(csvFile, pdfFile string) error {
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
                continue  // Skip the header row
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
        line.Color = plotutil.Color(i - 1) // Dynamically assign color
        points.Shape = draw.CircleGlyph{}  // Use draw.CircleGlyph
        p.Add(line, points)
        p.Legend.Add(fmt.Sprintf("Load %d", i), line, points)
    }

    p.X.Tick.Marker = plot.TimeTicks{Format: "15:04"}
    p.X.Tick.Length = vg.Points(10)
    p.Legend.Top = false

if err := p.Save(16*vg.Inch, 4*vg.Inch, pdfFile); err != nil {  // Increased width to 8 inches
    return err
}
    return nil
}

func plotWorkloadVolatilityIntervals(csvFile, outputFile string) error {
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
    p.Title.Text = "Workload Volatility Over Time"
    p.X.Label.Text = "Time"
    p.Y.Label.Text = "Volatility"

    for i := 1; i < len(records[0]); i++ { // Iterate over each workload
        var pts plotter.XYs
        for j, record := range records {
            if j == 0 { // Skip the header
                continue
            }
            t, err := time.Parse(time.RFC3339, record[0]) // Parse the time
            if err != nil {
                return err
            }
            x := float64(t.Unix())
            y, err := strconv.ParseFloat(record[i], 64) // Parse the workload volatility
            if err != nil {
                return err
            }
            pts = append(pts, plotter.XY{X: x, Y: y})
        }

        line, points, err := plotter.NewLinePoints(pts)
        if err != nil {
            return err
        }
        line.Color = plotutil.Color(i - 1)
        points.Shape = draw.CircleGlyph{}
        p.Add(line, points)
        p.Legend.Add(fmt.Sprintf("Volatility Load %d", i), line, points)
    }

    p.X.Tick.Marker = plot.TimeTicks{Format: "15:04"}
    p.X.Tick.Length = vg.Points(10)

    if err := p.Save(18*vg.Inch, 6*vg.Inch, outputFile); err != nil {
        return err
    }

    return nil
}

func plotWorkloadChanges(csvFile, outputFile string) error {
    // Open CSV file
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

    // Organize data by workload
    workloadData := make(map[string]plotter.XYs)
    for _, record := range records[1:] { // Skipping header
        timestamp, err := time.Parse(time.RFC3339, record[0])
        if err != nil {
            return err
        }
        change, err := strconv.ParseFloat(record[2], 64)
        if err != nil {
            return err
        }
        workload := record[1]
        workloadData[workload] = append(workloadData[workload], plotter.XY{
            X: float64(timestamp.Unix()),
            Y: change,
        })
    }

    // Create a plot
    p := plot.New()
    if err != nil {
        return err
    }
    p.Title.Text = "Workload Changes Over Time"
    p.X.Label.Text = "Time"
    p.Y.Label.Text = "Change"
    p.X.Tick.Marker = plot.TimeTicks{Format: "15:04"}

    // Define and use a color palette
    colorPalette := []color.Color{
        color.RGBA{R: 255, G: 0, B: 0, A: 255}, // Red
        color.RGBA{G: 255, B: 0, A: 255},       // Green
        color.RGBA{B: 255, A: 255},             // Blue
        // ... additional colors ...
    }

    // Plot each workload with a unique color
    colorIndex := 0
    for workload, data := range workloadData {
        line, points, err := plotter.NewLinePoints(data)
        if err != nil {
            return err
        }

        line.Color = colorPalette[colorIndex%len(colorPalette)]
        points.Shape = draw.CircleGlyph{}
        p.Add(line, points)
        p.Legend.Add(workload, line, points)

        colorIndex++
    }

    // Save the plot to a file
    if err := p.Save(24*vg.Inch, 8*vg.Inch, outputFile); err != nil {
        return err
    }

    return nil
}


func getUniqueWorkloads(csvFile string) ([]string, error) {
    f, err := os.Open(csvFile)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    r := csv.NewReader(f)
    records, err := r.ReadAll()
    if err != nil {
        return nil, err
    }

    uniqueWorkloads := make(map[string]struct{})
    for _, record := range records[1:] { // Skipping header
        workload := record[1]
        uniqueWorkloads[workload] = struct{}{}
    }

    var workloads []string
    for workload := range uniqueWorkloads {
        workloads = append(workloads, workload)
    }
    return workloads, nil
}

func generateColor(index int) color.Color {
    // Base colors
    baseColors := []color.RGBA{
        {R: 255, G: 0, B: 0, A: 255}, // Red
        {G: 255, B: 0, A: 255},       // Green
        {B: 255, A: 255},             // Blue
        {R: 255, G: 255, A: 255},     // Yellow
        {R: 255, B: 255, A: 255},     // Magenta
        {G: 255, B: 255, A: 255},     // Cyan
    }

    // Select a base color and modify it slightly
    base := baseColors[index%len(baseColors)]
    return color.RGBA{
        R: base.R + uint8(rand.Intn(55)),
        G: base.G + uint8(rand.Intn(55)),
        B: base.B + uint8(rand.Intn(55)),
        A: 255,
    }
}
