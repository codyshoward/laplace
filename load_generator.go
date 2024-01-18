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
    Load1          []float64 `json:"load1"`
    Load2          []float64 `json:"load2"`
    Load3          []float64 `json:"load3"`
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
        Load1:          generateRandomSlice(numIntegers, maxRange, isVolatile),
        Load2:          generateRandomSlice(numIntegers, maxRange, isVolatile),
        Load3:          generateRandomSlice(numIntegers, maxRange, isVolatile),
        ValueGenerated: valueGenerated,
    }
}


// Updated generateRandomSlice function
func generateRandomSlice(n int, maxRange float64, isVolatile bool) []float64 {
    slice := make([]float64, n)
    for i := 0; i < n; i++ {
        baseValue := rand.Float64() * maxRange
        if isVolatile {
            // Add volatility: modify the base value randomly
            volatilityFactor := rand.NormFloat64() * maxRange / 2
            slice[i] = baseValue + volatilityFactor
        } else {
            slice[i] = baseValue
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
