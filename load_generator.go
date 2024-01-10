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

func NewWorkload(name string, numIntegers int) *Workload {
    return &Workload{
        Name:           name,
        Load1:          generateRandomSlice(numIntegers),
        Load2:          generateRandomSlice(numIntegers),
        Load3:          generateRandomSlice(numIntegers),
        ValueGenerated: rand.Intn(100) + 1, // Random number between 1 and 100
    }
}

func generateRandomSlice(n int) []float64 {
    slice := make([]float64, n)
    for i := 0; i < n; i++ {
        // Generate a random float64 number. You can adjust the range as needed.
        slice[i] = rand.Float64() * 100 // Random float64 number between 0 and 100
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
        rand.Seed(time.Now().UnixNano())

        var numWorkloads, numIntegers int
        fmt.Print("Enter the number of workloads: ")
        fmt.Scanln(&numWorkloads)
        fmt.Print("Enter the number of integers in each load: ")
        fmt.Scanln(&numIntegers)

        workloads := make([]Workload, numWorkloads)
        for i := 0; i < numWorkloads; i++ {
                name := fmt.Sprintf("Workload%d", i+1)
                workload := NewWorkload(name, numIntegers)
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
