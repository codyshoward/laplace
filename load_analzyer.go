package main

// Import necessary packages for handling different functionalities
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
    "strconv"
)

// Workload struct represents the data structure for a workload
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

// Data struct is a container for a slice of Workloads
type Data struct {
    Workloads []Workload `json:"workloads"`
}

// TimedValue struct represents a value with an associated timestamp
type TimedValue struct {
    Timestamp time.Time `json:"timestamp"`
    Value     float64   `json:"value"`
}

// SummedWorkload struct aggregates the total loads for each timestamp
type SummedWorkload struct {
    Timestamp   time.Time
    TotalLoad1  float64
    TotalLoad2  float64
    TotalLoad3  float64
}

// main is the entry point of the application.
// It performs a series of operations to process and analyze workload data:
func main() {
    // Read the current directory to find workload files.
    files, err := os.ReadDir(".") 
    if err != nil {
        log.Fatalf("Error reading directory: %v", err)
    }

    var data Data // Initialize a Data struct to hold all the workload data.
    var errExport error // Variable to capture any errors during CSV export

    // Iterate over each file in the directory.
    for _, file := range files {
         // Check if the file name indicates a workload JSON file.
        if strings.HasPrefix(file.Name(), "Workload") && strings.HasSuffix(file.Name(), ".json") {
            // Load the workload data from the JSON file.
            loadedData, err := LoadData(file.Name()) // Load the data from the file
            if err != nil { 
                log.Printf("Error loading data from file %s: %v", file.Name(), err)
                continue // Skip to the next file on error.
            }
            // Append the loaded workloads to the main data struct.
            data.Workloads = append(data.Workloads, loadedData.Workloads...)
        }
    }

    // Calculate various statistics for the loaded workload data.
    CalculateWorkloadStats(&data)

    // Aggregate workloads data into a summarized form.
    summedWorkloads, err := aggregateWorkloads(data.Workloads)
    if err != nil {
        log.Printf("Error aggregating workloads: %v", err)
        return // Exit if aggregation fails.
    }

    // Export the aggregated workload data to a CSV file.
    errExport = exportWorkloadToCSV(summedWorkloads, "output.csv")
    if errExport != nil {
        log.Printf("Error exporting data to CSV: %v", errExport)
    }
    // Determine the peak usage among all workloads.
    peakUsage := findPeakUsage(data.Workloads)

    // Display the peak usage information.
    fmt.Printf("\nPeak Usage Information:\n")
    fmt.Printf("Timestamp of Peak Usage: %v\n", peakUsage.Timestamp)
    fmt.Printf("Total Usage at Peak: %.2f\n", peakUsage.TotalUsage)

    // Calculate and record the contributions of each workload at the peak usage.
    var contributions []WorkloadContribution
    for _, workload := range data.Workloads {
        totalLoadAtPeak := getLoadAtTimestamp(workload.Load1, peakUsage.Timestamp) +
                          getLoadAtTimestamp(workload.Load2, peakUsage.Timestamp) +
                          getLoadAtTimestamp(workload.Load3, peakUsage.Timestamp)

        contributions = append(contributions, WorkloadContribution{
            Name: workload.Name,
            LoadAtPeak: totalLoadAtPeak,
        })
    }

    // Sort the workloads based on their load contribution at the peak time.
    sort.Slice(contributions, func(i, j int) bool {
        return contributions[i].LoadAtPeak > contributions[j].LoadAtPeak
    })

    // Identify the top 10% contributors at the peak usage.
    topTenPercentIndex := len(contributions) / 10
    topContributors := contributions[:topTenPercentIndex]

    // Display the top contributing workloads.
    fmt.Println("\nTop 10% Workloads Contributing to Peak Usage:")
    for _, contributor := range topContributors {
        fmt.Printf("Workload: %s, Load at Peak: %.2f\n", contributor.Name, contributor.LoadAtPeak)
    }
    
    // Calculate and sort workloads based on their volatility.
    var volatilities []WorkloadVolatility
    for _, workload := range data.Workloads {
        vol1, vol2, vol3, _ := CalculateRelativeVolatility(workload, 5*time.Minute)
        avgVolatility := (vol1 + vol2 + vol3) / 3
        volatilities = append(volatilities, WorkloadVolatility{Name: workload.Name, Volatility: avgVolatility})
    }
    sort.Slice(volatilities, func(i, j int) bool {
        return volatilities[i].Volatility > volatilities[j].Volatility
    })

    // Divide the workloads into three categories based on their volatility.
    highVolatility := volatilities[:len(volatilities)/3]
    mediumVolatility := volatilities[len(volatilities)/3 : 2*len(volatilities)/3]
    lowVolatility := volatilities[2*len(volatilities)/3:]

    // Display the workloads in each volatility category.
    fmt.Println("\nHigh Volatility Workloads:")
    for _, workload := range highVolatility {
        fmt.Println(workload.Name)
    }

    fmt.Println("\nMedium Volatility Workloads:")
    for _, workload := range mediumVolatility {
        fmt.Println(workload.Name)
    }

    fmt.Println("\nLow Volatility Workloads:")
    for _, workload := range lowVolatility {
        fmt.Println(workload.Name)
    }
    
    // Write volatility data to a CSV file.
    err = writeVolatilityToFile("output.csv", "volatility_output.csv")
    if err != nil {
        panic(err)
    }
    // Write individual workload volatility data to a CSV file.
    err = WriteWorkloadVolatilityToFile(&data, "workload_volatility.csv") // Pass a pointer to data
    if err != nil {
        panic(err)
    }
    // Write workload volatility intervals to a CSV file.
    err = WriteWorkloadIntervalVolatilityToFile(&data, "workload_volatility_intervals.csv") // Pass a pointer to data
    if err != nil {
       panic(err)
        }
}


// processFile processes a single workload data file.
// It loads the data, calculates statistics, and can optionally print these statistics.
func processFile(filename string) {
     // Load the workload data from the specified file.
    data, err := LoadData(filename)
    if err != nil {
        log.Printf("Error loading data from file %s: %v", filename, err)
        return // Exit the function if data loading fails.
    }

    // Calculate various statistics for the loaded workload data.
    CalculateWorkloadStats(data)

    // Print the calculated statistics
    //PrintWorkloadStats(*data) //HEY LISTEN!! UNCOMMENT FME FOR TSHOOTING
}

// LoadData loads workload data from a JSON file and returns it as a Data struct.
func LoadData(filename string) (*Data, error) {
    var data Data

    // Open the file for reading.
    file, err := os.Open(filename)
    if err != nil {
        return nil, err // Return an error if file opening fails.
    }
    defer file.Close() // Ensure the file is closed after the function execution.

     // Decode the JSON data into the Data struct.
    jsonDecoder := json.NewDecoder(file)
    if err := jsonDecoder.Decode(&data); err != nil {
        return nil, err // Return an error if JSON decoding fails.
    }

    return &data, nil // Return the loaded data.
}

// CalculateWorkloadStats calculates various statistics for the workload data.
func CalculateWorkloadStats(data *Data) (float64, float64, float64, float64, float64) {
    // Initialize variables to hold cumulative statistics.
    var totalLoad1, totalLoad2, totalLoad3, totalCost, totalValueGenerated float64

    // Iterate over each workload to sum up loads and value generated.
    for i := range data.Workloads {
        workload := &data.Workloads[i]

        // Normalize the lengths of Load1, Load2, and Load3 to ensure consistency.
        normalizeLoadLengths(&workload.Load1, &workload.Load2, &workload.Load3)
        
        // Calculate and store the total load for each type.
        workload.TotalLoad1 = TimedValue{Value: sum(workload.Load1)}
        workload.TotalLoad2 = TimedValue{Value: sum(workload.Load2)}
        workload.TotalLoad3 = TimedValue{Value: sum(workload.Load3)}

        // Calculate the total cost for the workload.
        workload.TotalCost = calculateTotalCost(workload)

        // Accumulate totals across all workloads.
        totalLoad1 += workload.TotalLoad1.Value
        totalLoad2 += workload.TotalLoad2.Value
        totalLoad3 += workload.TotalLoad3.Value
        totalValueGenerated += workload.ValueGenerated
    }
    
    // Calculate the grand total cost.
    totalCost = totalLoad1 + totalLoad2 + totalLoad3

    // Sum of all loads.
    totalLoadSum := totalLoad1 + totalLoad2 + totalLoad3

    // Variables for grand total loads.
    var grandTotalLoad1, grandTotalLoad2, grandTotalLoad3 float64

    // Calculate grand totals for each load type.
    CalculateGrandSums(data, &grandTotalLoad1, &grandTotalLoad2, &grandTotalLoad3)

    // Calculate the average load across all workloads.
    averageTotalLoad := totalLoadSum / float64(len(data.Workloads))

    // Variables for accumulating deviation sums.
    var upwardDevSum, downwardDevSum float64

    // Calculate relative contributions and deviations for each workload.
    for _, workload := range data.Workloads {
        calculateRelativeContributionsAndDeviations(&workload, grandTotalLoad1, grandTotalLoad2, grandTotalLoad3, totalValueGenerated, averageTotalLoad, totalCost, totalLoadSum, &upwardDevSum, &downwardDevSum)
    }
    
    // Return cumulative statistics.
    return totalLoad1, totalLoad2, totalLoad3, totalCost, totalValueGenerated
}

// calculateRelativeContributionsAndDeviations calculates and sets relative contribution and deviation values for a workload.
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
    // Calculate volatilities
    volatilityLoad1, volatilityLoad2, volatilityLoad3, err := CalculateRelativeVolatility(*workload, 5 * time.Minute)
    if err != nil {
        fmt.Printf("Error calculating volatility for workload %s: %v\n", workload.Name, err)
        return
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
    fmt.Printf("  Volatility Load 1: %.2f\n", volatilityLoad1)
    fmt.Printf("  Volatility Load 2: %.2f\n", volatilityLoad2)
    fmt.Printf("  Volatility Load 3: %.2f\n", volatilityLoad3)
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
func CalculateRelativeVolatility(workload Workload, interval time.Duration) (float64, float64, float64, error) {
    load1Averages, err := calculateIntervalAverages(workload.Load1, interval)
    if err != nil {
        return 0, 0, 0, err
    }
    load2Averages, err := calculateIntervalAverages(workload.Load2, interval)
    if err != nil {
        return 0, 0, 0, err
    }
    load3Averages, err := calculateIntervalAverages(workload.Load3, interval)
    if err != nil {
        return 0, 0, 0, err
    }

    volatilityLoad1 := calculateStandardDeviation(load1Averages)
    volatilityLoad2 := calculateStandardDeviation(load2Averages)
    volatilityLoad3 := calculateStandardDeviation(load3Averages)

    return volatilityLoad1, volatilityLoad2, volatilityLoad3, nil
}

func calculateIntervalAverages(timedValues []TimedValue, interval time.Duration) ([]float64, error) {
    if len(timedValues) == 0 {
        return nil, fmt.Errorf("timedValues is empty")
    }

    var intervalAverages []float64
    var sum float64
    var count int
    startTime := timedValues[0].Timestamp

    for _, timedValue := range timedValues {
        if timedValue.Timestamp.Sub(startTime) <= interval {
            sum += timedValue.Value
            count++
        } else {
            intervalAverages = append(intervalAverages, sum/float64(count))
            sum = timedValue.Value
            count = 1
            startTime = timedValue.Timestamp
        }
    }
    // Add the last interval's average if any data is left
    if count > 0 {
        intervalAverages = append(intervalAverages, sum/float64(count))
    }

    return intervalAverages, nil
}

func calculateStandardDeviation(values []float64) float64 {
    if len(values) == 0 {
        return 0
    }

    mean := 0.0
    for _, v := range values {
        mean += v
    }
    mean /= float64(len(values))

    variance := 0.0
    for _, v := range values {
        variance += (v - mean) * (v - mean)
    }
    variance /= float64(len(values))

    return math.Sqrt(variance)
}
type PeakUsage struct {
    Timestamp time.Time
    TotalUsage float64
}

func findPeakUsage(workloads []Workload) PeakUsage {
    usageMap := make(map[time.Time]float64)

    for _, workload := range workloads {
        for _, load := range workload.Load1 {
            usageMap[load.Timestamp] += load.Value
        }
        for _, load := range workload.Load2 {
            usageMap[load.Timestamp] += load.Value
        }
        for _, load := range workload.Load3 {
            usageMap[load.Timestamp] += load.Value
        }
    }

    var peakUsage PeakUsage
    for timestamp, totalUsage := range usageMap {
        if totalUsage > peakUsage.TotalUsage {
            peakUsage = PeakUsage{
                Timestamp: timestamp,
                TotalUsage: totalUsage,
            }
        }
    }

    return peakUsage
}
func getLoadAtTimestamp(timedValues []TimedValue, timestamp time.Time) float64 {
    for _, tv := range timedValues {
        if tv.Timestamp == timestamp {
            return tv.Value
        }
    }
    return 0
}
type WorkloadContribution struct {
    Name       string
    LoadAtPeak float64
}

type WorkloadVolatility struct {
    Name       string
    Volatility float64
}

func writeVolatilityToFile(csvInputFile, outputFile string) error {
    f, err := os.Open(csvInputFile)
    if err != nil {
        return err
    }
    defer f.Close()

    r := csv.NewReader(f)
    records, err := r.ReadAll()
    if err != nil {
        return err
    }

    outFile, err := os.Create(outputFile)
    if err != nil {
        return err
    }
    defer outFile.Close()

    writer := csv.NewWriter(outFile)
    defer writer.Flush()

    // Write header
    header := []string{"Time"}
    for i := 1; i < len(records[0]); i++ {
        header = append(header, fmt.Sprintf("Volatility Load %d", i))
    }
    if err := writer.Write(header); err != nil {
        return err
    }

    // Calculate and write volatilities
    for i := 5; i < len(records); i += 5 {
        var row []string
        row = append(row, records[i][0]) // Add timestamp
        for j := 1; j < len(records[0]); j++ {
            volatility, err := calculateVolatilityAtInterval(records, j, i)
            if err != nil {
                return err
            }
            row = append(row, fmt.Sprintf("%.2f", volatility))
        }
        if err := writer.Write(row); err != nil {
            return err
        }
    }

    return nil
}



func calculateVolatilityAtInterval(records [][]string, column, interval int) (float64, error) {
    var values []float64

    // Calculate the start index for the 5-minute interval
    start := interval - 5
    if start < 0 {
        start = 0
    }

    // Collect values for the interval
    for i := start; i < interval && i < len(records); i++ {
        value, err := strconv.ParseFloat(records[i][column], 64)
        if err != nil {
            return 0.0, err
        }
        values = append(values, value)
    }

    return calculateStandardDeviation(values), nil
}


func calculateStandardDeviationIntervals(values []float64) float64 {
    if len(values) == 0 {
        return 0.0
    }

    mean := 0.0
    for _, v := range values {
        mean += v
    }
    mean /= float64(len(values))

    variance := 0.0
    for _, v := range values {
        variance += (v - mean) * (v - mean)
    }
    variance /= float64(len(values) - 1) // Use (N-1) for sample standard deviation

    return math.Sqrt(variance)
}

func CalculateWorkloadIntervalSums(data *Data) map[string][]TimedValue {
    intervalSums := make(map[string][]TimedValue)

    for _, workload := range data.Workloads {
        var intervalSum []TimedValue

        // Assuming the loads are in chronological order and each represents a minute
        for i := 0; i < len(workload.Load1); i += 5 {
            sum := 0.0

            // Calculate sum for the interval
            for j := i; j < i+5 && j < len(workload.Load1); j++ {
                sum += workload.Load1[j].Value + workload.Load2[j].Value + workload.Load3[j].Value
            }

            // Add interval sum to the list
            intervalSum = append(intervalSum, TimedValue{
                Timestamp: workload.Load1[i].Timestamp,
                Value:     sum,
            })
        }

        intervalSums[workload.Name] = intervalSum
    }

    return intervalSums
}

func calculateLoadSumInIntervals(workload Workload, interval int) ([]float64, error) {
    var sums []float64
    for i := 0; i < len(workload.Load1); i += interval {
        sum := 0.0
        count := 0
        for j := i; j < i+interval && j < len(workload.Load1); j++ {
            sum += workload.Load1[j].Value + workload.Load2[j].Value + workload.Load3[j].Value
            count++
        }
        if count > 0 {
            sums = append(sums, sum/float64(count)) // Average sum for this interval
        }
    }
    return sums, nil
}

func calculateVolatility(sums []float64) float64 {
    mean, variance := 0.0, 0.0
    for _, sum := range sums {
        mean += sum
    }
    mean /= float64(len(sums))

    for _, sum := range sums {
        variance += (sum - mean) * (sum - mean)
    }
    variance /= float64(len(sums))

    return math.Sqrt(variance)
}

func WriteWorkloadVolatilityToFile(data *Data, outputFile string) error {
    file, err := os.Create(outputFile)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write the header
    if err := writer.Write([]string{"Workload", "Volatility"}); err != nil {
        return err
    }

    for _, workload := range data.Workloads {
        sums, err := calculateLoadSumInIntervals(workload, 5) // 5-minute intervals
        if err != nil {
            return err
        }

        volatility := calculateVolatility(sums)

        record := []string{
            workload.Name,
            fmt.Sprintf("%.2f", volatility),
        }

        if err := writer.Write(record); err != nil {
            return err
        }
    }

    return nil
}
func calculateIntervalSumsWithTimestamps(workload Workload, intervalSize int) ([]TimedValue, error) {
    var intervalSums []TimedValue

    for i := 0; i < len(workload.Load1); i += intervalSize {
        sum := 0.0
        count := 0
        var timestamp time.Time

        for j := i; j < i+intervalSize && j < len(workload.Load1); j++ {
            sum += workload.Load1[j].Value + workload.Load2[j].Value + workload.Load3[j].Value
            timestamp = workload.Load1[j].Timestamp
            count++
        }

        if count > 0 {
            avgSum := sum / float64(count)
            intervalSums = append(intervalSums, TimedValue{Timestamp: timestamp, Value: avgSum})
        }
    }

    return intervalSums, nil
}

func calculateVolatilityIntervals(intervalSums []TimedValue) float64 {
    var values []float64
    for _, timedValue := range intervalSums {
        values = append(values, timedValue.Value)
    }
    return calculateStandardDeviation(values)
}
func WriteWorkloadIntervalVolatilityToFile(data *Data, outputFile string) error {
    file, err := os.Create(outputFile)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    if err := writer.Write([]string{"Timestamp", "Workload", "Change"}); err != nil {
        return err
    }

    for _, workload := range data.Workloads {
        intervalChanges, err := calculateIntervalSumChanges(workload, 5) // 5-minute intervals
        if err != nil {
            return err
        }

        for _, intervalChange := range intervalChanges {
            record := []string{
                intervalChange.Timestamp.Format(time.RFC3339),
                workload.Name,
                fmt.Sprintf("%.2f", intervalChange.Value),
            }
            if err := writer.Write(record); err != nil {
                return err
            }
        }
    }

    return nil
}

func calculateIntervalSumChanges(workload Workload, intervalSize int) ([]TimedValue, error) {
    var intervalSums []TimedValue

    var previousSum float64
    for i := 0; i < len(workload.Load1); i += intervalSize {
        sum := 0.0
        var timestamp time.Time

        for j := i; j < i+intervalSize && j < len(workload.Load1); j++ {
            sum += workload.Load1[j].Value + workload.Load2[j].Value + workload.Load3[j].Value
            timestamp = workload.Load1[j].Timestamp
        }

        // Calculate change from the previous interval
        change := sum - previousSum
        if i != 0 { // Skip the first interval as it has no previous data
            intervalSums = append(intervalSums, TimedValue{Timestamp: timestamp, Value: change})
        }
        previousSum = sum
    }

    return intervalSums, nil
}
