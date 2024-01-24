package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "math"
    "strings"
    "time"
    "sort"
    "encoding/csv"
)

type Workload struct {
    Name                  string    `json:"name"`
    Load1                 []TimedValue `json:"load1"`
    Load2                 []TimedValue `json:"load2"`
    Load3                 []TimedValue `json:"load3"`
    ValueGenerated        float64   `json:"valueGenerated"`
    TotalLoad1            TimedValue
    TotalLoad2            TimedValue
    TotalLoad3            TimedValue
    RelativeValueGenerated float64
    RelativeLoad1         float64
    RelativeLoad2         float64
    RelativeLoad3         float64
    TotalLoad             float64
    TotalRelativeLoad     float64
    RelativeCost          float64
    TotalCost             float64
    Timestamp             time.Time
}

type Data struct {
    Workloads []Workload `json:"workloads"`
}

type TimedValue struct {
    Timestamp time.Time
    Value     float64
}

type SummedWorkload struct {
    Timestamp   time.Time
    TotalLoad1  float64
    TotalLoad2  float64
    TotalLoad3  float64
}

func main() {
    files, err := os.ReadDir(".")
    if err != nil {
        log.Fatalf("Error reading directory: %v", err)
    }

    var data Data // Define the data variable
    var errExport error // Declare errExport for exportWorkloadToCSV

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

    // Aggregate workloads into summed workloads
    summedWorkloads, err := aggregateWorkloads(data.Workloads)
    if err != nil {
        log.Printf("Error aggregating workloads: %v", err)
        return
    }

    // Call exportWorkloadToCSV to export the data to a CSV file
    errExport = exportWorkloadToCSV(summedWorkloads, "output.csv")
    if errExport != nil {
        log.Printf("Error exporting data to CSV: %v", errExport)
    }
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

    // Sum up all the loads and the value generated across all workloads
    for i := range data.Workloads {
        workload := &data.Workloads[i]

         // Normalize the lengths of Load1, Load2, Load3
        normalizeLoadLengths(&workload.Load1, &workload.Load2, &workload.Load3)

        workload.TotalLoad1 = TimedValue{Value: sum(workload.Load1)}
        workload.TotalLoad2 = TimedValue{Value: sum(workload.Load2)}
        workload.TotalLoad3 = TimedValue{Value: sum(workload.Load3)}

        //workload.TotalCost = calculateTotalCost(workload.TotalLoad1.Value, workload.TotalLoad2.Value, workload.TotalLoad3.Value)
        workload.TotalCost = calculateTotalCost(workload)

        totalLoad1 += workload.TotalLoad1.Value
        totalLoad2 += workload.TotalLoad2.Value
        totalLoad3 += workload.TotalLoad3.Value
        totalValueGenerated += workload.ValueGenerated
    }

    totalCost = totalLoad1 + totalLoad2 + totalLoad3

    // Total load sum
    totalLoadSum := totalLoad1 + totalLoad2 + totalLoad3

    // Initialize grandTotalLoadX variables
    var grandTotalLoad1, grandTotalLoad2, grandTotalLoad3 float64

    // Call CalculateGrandSums to calculate the grand totals
    CalculateGrandSums(data, &grandTotalLoad1, &grandTotalLoad2, &grandTotalLoad3)

    // Calculate the average total load
    averageTotalLoad := totalLoadSum / float64(len(data.Workloads))

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
        workload.RelativeLoad1 = (workload.TotalLoad1.Value / grandTotalLoad1) * 100
    }
    if grandTotalLoad2 > 0 {
        workload.RelativeLoad2 = (workload.TotalLoad2.Value / grandTotalLoad2) * 100
    }
    if grandTotalLoad3 > 0 {
        workload.RelativeLoad3 = (workload.TotalLoad3.Value / grandTotalLoad3) * 100
    }
    if totalLoadSum > 0 {
        workload.RelativeCost = (workload.TotalCost / totalCost) * 100
    }

    // Calculate relative value generated
    if totalValueGenerated > 0 {
        workload.RelativeValueGenerated = (workload.ValueGenerated / totalValueGenerated) * 100
    }

    // Calculate deviations
    totalLoad := workload.TotalLoad1.Value + workload.TotalLoad2.Value + workload.TotalLoad3.Value
    deviation := totalLoad - averageTotalLoad
    if deviation > 0 {
        *upwardDevSum += deviation * deviation
    } else {
        *downwardDevSum += deviation * deviation
    }

    // Print workload statistics
    fmt.Printf("Workload: %s\n", workload.Name)
    fmt.Printf("  Total Load 1: %.2f\n", workload.TotalLoad1.Value)
    fmt.Printf("  Total Load 2: %.2f\n", workload.TotalLoad2.Value)
    fmt.Printf("  Total Load 3: %.2f\n", workload.TotalLoad3.Value)
    fmt.Printf("  Relative Load 1: %.2f%%\n", workload.RelativeLoad1)
    fmt.Printf("  Relative Load 2: %.2f%%\n", workload.RelativeLoad2)
    fmt.Printf("  Relative Load 3: %.2f%%\n", workload.RelativeLoad3)
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
func sum(timedValues []TimedValue) float64 {
    var total float64
    for _, timedValue := range timedValues {
        total += timedValue.Value
    }
    return total
}
//We sum all the loads in the individual workloads togethor to get a sum that we can then use to get relative value
func CalculateGrandSums(data *Data, grandTotalLoad1, grandTotalLoad2, grandTotalLoad3 *float64) {
    for _, workload := range data.Workloads {
        *grandTotalLoad1 += workload.TotalLoad1.Value
        *grandTotalLoad2 += workload.TotalLoad2.Value
        *grandTotalLoad3 += workload.TotalLoad3.Value
    }
}

func calculateTotalCost(workload *Workload) float64 {
    return workload.TotalLoad1.Value + workload.TotalLoad2.Value + workload.TotalLoad3.Value
}

func normalizeLoadLengths(load1, load2, load3 *[]TimedValue) {
    //maxLength := int(maxOfThree(float64(len(*load1)), float64(len(*load2)), float64(len(*load3))))
    maxLength := maxOfThree(len(*load1), len(*load2), len(*load3))

    *load1 = fillMissingValues(*load1, maxLength)
    *load2 = fillMissingValues(*load2, maxLength)
    *load3 = fillMissingValues(*load3, maxLength)
}

func fillMissingValues(load []TimedValue, length int) []TimedValue {
    if len(load) == length {
        return load
    }

    averageValue := average(load)
    for len(load) < length {
        load = append(load, TimedValue{Value: averageValue})
    }

    return load
}

func average(load []TimedValue) float64 {
    if len(load) == 0 {
        return 0
    }

    sum := 0.0
    for _, val := range load {
        sum += val.Value
    }
    return sum / float64(len(load))
}
func maxOfThree(a, b, c int) int {
    maxVal := a
    if b > maxVal {
        maxVal = b
    }
    if c > maxVal {
        maxVal = c
    }
    return maxVal
}

// aggregateWorkloads aggregates load values for each unique timestamp across all workloads.
func aggregateWorkloads(workloads []Workload) ([]SummedWorkload, error) {
    summedWorkloadsMap := make(map[time.Time]SummedWorkload)

    for _, workload := range workloads {
        for _, load1 := range workload.Load1 {
            summedWorkload := summedWorkloadsMap[load1.Timestamp]
            summedWorkload.Timestamp = load1.Timestamp
            summedWorkload.TotalLoad1 += load1.Value
            summedWorkloadsMap[load1.Timestamp] = summedWorkload
        }
        for _, load2 := range workload.Load2 {
            summedWorkload := summedWorkloadsMap[load2.Timestamp]
            summedWorkload.TotalLoad2 += load2.Value
            summedWorkloadsMap[load2.Timestamp] = summedWorkload
        }
        for _, load3 := range workload.Load3 {
            summedWorkload := summedWorkloadsMap[load3.Timestamp]
            summedWorkload.TotalLoad3 += load3.Value
            summedWorkloadsMap[load3.Timestamp] = summedWorkload
        }
    }

    return sortSummedWorkloads(summedWorkloadsMap), nil
}

// sortSummedWorkloads converts a map of summed workloads to a sorted slice.
func sortSummedWorkloads(summedWorkloadsMap map[time.Time]SummedWorkload) []SummedWorkload {
    var summedWorkloads []SummedWorkload
    for _, workload := range summedWorkloadsMap {
        summedWorkloads = append(summedWorkloads, workload)
    }
    sort.Slice(summedWorkloads, func(i, j int) bool {
        return summedWorkloads[i].Timestamp.Before(summedWorkloads[j].Timestamp)
    })
    return summedWorkloads
}

// exportWorkloadToCSV exports the summed workload data to a CSV file.
func exportWorkloadToCSV(workloads []SummedWorkload, filename string) error {
    // CSV file creation and error handling remains the same.

    // Create a CSV file
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()
    // Write data
    for _, workload := range workloads {
        timestamp := workload.Timestamp.Format(time.RFC3339)
        totalLoad1 := fmt.Sprintf("%f", workload.TotalLoad1)
        totalLoad2 := fmt.Sprintf("%f", workload.TotalLoad2)
        totalLoad3 := fmt.Sprintf("%f", workload.TotalLoad3)

        if err := writer.Write([]string{timestamp, totalLoad1, totalLoad2, totalLoad3}); err != nil {
            return err
        }
    }

    return nil
}


func validateWorkloadSynchronization(workload Workload) error {
    if len(workload.Load1) != len(workload.Load2) || len(workload.Load1) != len(workload.Load3) {
        return fmt.Errorf("workload '%s' has unsynchronized load arrays", workload.Name)
    }
    return nil
}
