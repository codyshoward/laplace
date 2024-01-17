package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "math"
    "strings"
)

type Workload struct {
    Name                  string    `json:"name"`
    Load1                 []float64 `json:"load1"`
    Load2                 []float64 `json:"load2"`
    Load3                 []float64 `json:"load3"`
    ValueGenerated        float64   `json:"valueGenerated"`
    TotalLoad1            float64
    TotalLoad2            float64
    TotalLoad3            float64
    RelativeValueGenerated float64
    RelativeLoad1         float64
    RelativeLoad2         float64
    RelativeLoad3         float64
    TotalLoad             float64
    TotalRelativeLoad     float64
    RelativeCost          float64
    TotalCost             float64
    VolatilityLoad1       float64
    VolatilityLoad2       float64
    VolatilityLoad3       float64
}

type Data struct {
    Workloads []Workload `json:"workloads"`
}

func main() {
    files, err := os.ReadDir(".")
    if err != nil {
        log.Fatalf("Error reading directory: %v", err)
    }

    var data Data // Define the data variable

    for _, file := range files {
        if strings.HasPrefix(file.Name(), "Workload") && strings.HasSuffix(file.Name(), ".json") {
            loadedData, err := LoadData(file.Name()) // Load the data from the file
            if err != nil {
                log.Printf("Error loading data from file %s: %v", file.Name(), err)
                continue
            }
            data.Workloads = append(data.Workloads, loadedData.Workloads...)
        }
    }

    // Calculate statistics for the loaded data
    CalculateWorkloadStats(&data)

    // Print the calculated statistics
    //PrintWorkloadStats(data) //HEY LISTEN!! UNCOMMNET ME FOR TSHOOTING
}

func processFile(filename string) {
    data, err := LoadData(filename)
    if err != nil {
        log.Printf("Error loading data from file %s: %v", filename, err)
        return
    }

    // Calculate statistics for the loaded data
    CalculateWorkloadStats(data)

    // Print the calculated statistics
    //PrintWorkloadStats(*data) //HEY LISTEN!! UNCOMMENT FME FOR TSHOOTING
}

func LoadData(filename string) (*Data, error) {
    var data Data
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    jsonDecoder := json.NewDecoder(file)
    if err := jsonDecoder.Decode(&data); err != nil {
        return nil, err
    }

    return &data, nil
}

func CalculateWorkloadStats(data *Data) (float64, float64, float64, float64, float64) {
    var totalLoad1, totalLoad2, totalLoad3, totalCost, totalValueGenerated float64

    // First, sum up all the loads and the value generated across all workloads
    for i := range data.Workloads {
        workload := &data.Workloads[i]

        workload.TotalLoad1 = sum(workload.Load1)
        workload.TotalLoad2 = sum(workload.Load2)
        workload.TotalLoad3 = sum(workload.Load3)

        workload.TotalCost = calculateTotalCost(workload)

        totalLoad1 += workload.TotalLoad1
        totalLoad2 += workload.TotalLoad2
        totalLoad3 += workload.TotalLoad3
        totalValueGenerated += workload.ValueGenerated
        totalCost = (totalLoad1 + totalLoad2 + totalLoad3)
        workload.calculateVolatility()
    }

    // Total load sum
    totalLoadSum := totalLoad1 + totalLoad2 + totalLoad3

    // Initialize grandTotalLoadX variables
    var grandTotalLoad1, grandTotalLoad2, grandTotalLoad3 float64

    // Call CalculateGrandSums to calculate the grand totals
    CalculateGrandSums(data, &grandTotalLoad1, &grandTotalLoad2, &grandTotalLoad3)

    // Calculate the average total load
    averageTotalLoad := totalLoadSum / float64(len(data.Workloads))
    //averageLoad1 := totalLoad1 / float64(len(data.Workloads))
    //averageLoad2 := totalLoad2 / float64(len(data.Workloads))
    //averageLoad3 := totalLoad3 / float64(len(data.Workloads))

    // Variables to accumulate squared deviations
    var upwardDevSum, downwardDevSum float64

    // Calculate relative contributions and deviations
    for _, workload := range data.Workloads {
        calculateRelativeContributionsAndDeviations(&workload, grandTotalLoad1, grandTotalLoad2, grandTotalLoad3, totalValueGenerated, averageTotalLoad, totalCost, totalLoadSum, &upwardDevSum, &downwardDevSum)
    }

    return totalLoad1, totalLoad2, totalLoad3, totalCost, totalValueGenerated
}

func calculateRelativeContributionsAndDeviations(workload *Workload, grandTotalLoad1, grandTotalLoad2, grandTotalLoad3, totalValueGenerated, averageTotalLoad, totalCost, totalLoadSum float64, upwardDevSum, downwardDevSum *float64) {
    // Calculate relative loads
    if grandTotalLoad1 > 0 {
        workload.RelativeLoad1 = (workload.TotalLoad1 / grandTotalLoad1) * 100
    }
    if grandTotalLoad2 > 0 {
        workload.RelativeLoad2 = (workload.TotalLoad2 / grandTotalLoad2) * 100
    }
    if grandTotalLoad3 > 0 {
        workload.RelativeLoad3 = (workload.TotalLoad3 / grandTotalLoad3) * 100
    }
    if totalLoadSum > 0 {
        workload.RelativeCost = (workload.TotalCost / totalCost) * 100
    }

    // Calculate relative value generated
    if totalValueGenerated > 0 {
        workload.RelativeValueGenerated = (workload.ValueGenerated / totalValueGenerated) * 100
    }

    // Calculate deviations
    totalLoad := workload.TotalLoad1 + workload.TotalLoad2 + workload.TotalLoad3
    deviation := totalLoad - averageTotalLoad
    if deviation > 0 {
        *upwardDevSum += deviation * deviation
    } else {
        *downwardDevSum += deviation * deviation
    }

    // Print workload statistics
    fmt.Printf("Workload: %s\n", workload.Name)
    fmt.Printf("  Total Load 1: %.2f\n", workload.TotalLoad1)
    fmt.Printf("  Total Load 2: %.2f\n", workload.TotalLoad2)
    fmt.Printf("  Total Load 3: %.2f\n", workload.TotalLoad3)
    fmt.Printf("  Relative Load 1: %.2f%%\n", workload.RelativeLoad1)
    fmt.Printf("  Relative Load 2: %.2f%%\n", workload.RelativeLoad2)
    fmt.Printf("  Relative Load 3: %.2f%%\n", workload.RelativeLoad3)
    fmt.Printf("  Load 1 Volatility: %.2f%%\n", workload.VolatilityLoad1)
    fmt.Printf("  Load 2 Volatility: %.2f%%\n", workload.VolatilityLoad2)
    fmt.Printf("  Load 3 Volatility: %.2f%%\n", workload.VolatilityLoad3)
    fmt.Printf("  Total Cost: %.2f\n", workload.TotalCost)
    fmt.Printf("  Total Relative Load and Cost: %.2f%%\n", workload.RelativeCost)
    fmt.Printf("  Relative Value Generated: %.2f%%\n", workload.RelativeValueGenerated)
    // Add two empty lines for separation
    fmt.Println()
    fmt.Println()
}



func printStandardDeviations(workloadCount int, upwardDevSum, downwardDevSum float64) {
    if workloadCount > 0 {
        upwardStdDev := math.Sqrt(upwardDevSum / float64(workloadCount))
        downwardStdDev := math.Sqrt(downwardDevSum / float64(workloadCount))
        fmt.Printf("Upward Standard Deviation: %.2f\n", upwardStdDev)
        fmt.Printf("Downward Standard Deviation: %.2f\n", downwardStdDev)
    } else {
        fmt.Println("Upward Standard Deviation: N/A")
        fmt.Println("Downward Standard Deviation: N/A")
    }
}

//We slice the array up like a pizza
func sum(slice []float64) float64 {
    total := 0.0
    for _, value := range slice {
        total += value
    }
    return total
}
//We sum all the loads in the individual workloads togethor to get a sum that we can then use to get relative value
func CalculateGrandSums(data *Data, grandTotalLoad1, grandTotalLoad2, grandTotalLoad3 *float64) {
    for _, workload := range data.Workloads {
        *grandTotalLoad1 += sum(workload.Load1)
        *grandTotalLoad2 += sum(workload.Load2)
        *grandTotalLoad3 += sum(workload.Load3)
    }
}

func calculateTotalCost(workload *Workload) float64 {
    return workload.TotalLoad1 + workload.TotalLoad2 + workload.TotalLoad3
}
func (w *Workload) calculateVolatility() {
    loadSets := [][]float64{w.Load1, w.Load2, w.Load3}
    volatilities := []*float64{&w.VolatilityLoad1, &w.VolatilityLoad2, &w.VolatilityLoad3}

    for i, numbers := range loadSets {
        if len(numbers) == 0 {
            continue
        }

        sum := 0.0
        max := numbers[0]
        min := numbers[0]

        for _, number := range numbers {
            sum += number
            if number > max {
                max = number
            }
            if number < min {
                min = number
            }
        }

        avg := sum / float64(len(numbers))
        upwardVolatility := math.Round((max - avg) * 100 / avg)
        downwardVolatility := math.Round((avg - min) * 100 / avg)

        // Use the greater of the two volatilities
        if upwardVolatility > downwardVolatility {
            *volatilities[i] = float64(upwardVolatility)
        } else {
            *volatilities[i] = float64(downwardVolatility)
        }
    }
}
