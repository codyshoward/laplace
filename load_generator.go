package main

import (
    "encoding/json"
    "fmt"
    "math/rand"
    "os"
    "time"
)

type Workload struct {
    Name           string    `json:"name"`
    Load1          []TimedValue `json:"load1"`
    Load2          []TimedValue `json:"load2"`
    Load3          []TimedValue `json:"load3"`
    ValueGenerated int       `json:"valueGenerated"`
}

const (
    highVolatilityPercentage = 30 // 30% of the workloads will have high volatility
)

// Modified NewWorkload function
func NewWorkload(name string, numIntegers int, isVolatile bool) *Workload {
    var maxRange float64 = 100

    if isVolatile {
        maxRange = 200 // Higher range for volatile workloads
    }

    // Adjust ValueGenerated with potential for higher volatility
    valueGenerated := rand.Intn(100) + 1
    if isVolatile {
        valueGenerated += rand.Intn(200) // Adding more variability
    }

    return &Workload{
        Name:           name,
        Load1:          generateTimedRandomSlice(numIntegers, maxRange, isVolatile),
        Load2:          generateTimedRandomSlice(numIntegers, maxRange, isVolatile),
        Load3:          generateTimedRandomSlice(numIntegers, maxRange, isVolatile),
        ValueGenerated: valueGenerated,
    }
}


type TimedValue struct {
    Timestamp time.Time
    Value     float64
}

func generateTimedRandomSlice(n int, maxRange float64, isVolatile bool) []TimedValue {
    slice := make([]TimedValue, n)
    startTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)

    for i := 0; i < n; i++ {
        // Generate a random or volatile value
        generatedValue := rand.Float64() * maxRange
        if isVolatile {
            // Apply your volatility logic here
            generatedValue += rand.NormFloat64() * maxRange / 2
        }

        // Store the timestamp and value
        slice[i] = TimedValue{
            Timestamp: startTime.Add(time.Duration(i) * time.Minute),
            Value:     generatedValue,
        }
    }
    return slice
}




func (w *Workload) SerializeToJson() ([]byte, error) {
    data, err := json.MarshalIndent(struct {
        Workloads []Workload `json:"workloads"`
    }{Workloads: []Workload{*w}}, "", "    ")
    if err != nil {
        return nil, err
    }
    return data, nil
}


func main() {
    rand.Seed(time.Now().UnixNano())

    // User input for number of workloads and integers per load
    var numWorkloads, numIntegers int
    fmt.Print("Enter the number of workloads: ")
    fmt.Scanln(&numWorkloads)
    fmt.Print("Enter the number of integers in each load: ")
    fmt.Scanln(&numIntegers)

    // Initialize workloads slice
    workloads := make([]Workload, numWorkloads)

    // Determine counts for less volatile and volatile workloads
    numLessVolatile := int(float64(numWorkloads) * 30 / 100)
    numVolatile := numWorkloads - numLessVolatile

    // Create a slice of indices and shuffle it
    indices := make([]int, numWorkloads)
    for i := range indices {
        indices[i] = i
    }
    rand.Shuffle(len(indices), func(i, j int) {
        indices[i], indices[j] = indices[j], indices[i]
    })

    // Assign volatility based on shuffled indices
    isVolatileMap := make(map[int]bool)
    for _, idx := range indices[:numVolatile] {
        isVolatileMap[idx] = true
    }

    // Generate workloads with specified volatility
    for i := 0; i < numWorkloads; i++ {
        name := fmt.Sprintf("Workload%d", i+1)
        isVolatile := isVolatileMap[i]
        workload := NewWorkload(name, numIntegers, isVolatile)
        workloads[i] = *workload
    }

        for _, workload := range workloads {
                data, err := workload.SerializeToJson()
                if err != nil {
                        fmt.Println("Error marshaling data:", err)
                        return
                }

                // Writing JSON data to a file
                file, err := os.Create(fmt.Sprintf("%s.json", workload.Name))
                if err != nil {
                        fmt.Println("Error creating file:", err)
                        return
                }
                defer file.Close()

                _, err = file.Write(data)
                if err != nil {
                        fmt.Println("Error writing to file:", err)
                        return
                }

                fmt.Printf("Workload %s generated and saved in %s.json\n", workload.Name, workload.Name)
        }
}
func generateWorkloads(numWorkloads, numIntegers int) []Workload {
    var workloads []Workload

    // Determine counts for less volatile and volatile workloads
    numLessVolatile := int(float64(numWorkloads) * 30 / 100)

    for i := 0; i < numWorkloads; i++ {
        // Determine the volatility for the workload
        isVolatile := i >= numLessVolatile // First 30% will be less volatile

        workload := NewWorkload(fmt.Sprintf("Workload%d", i+1), numIntegers, isVolatile)
        workloads = append(workloads, *workload)
    }

    return workloads
}
