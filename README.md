Welcome to Laplace, a tool for observing and balancing work. 
Created and maintined by Cody S Howard. Contact: codyshoward@gmail.com

Update: 01/10/2024
Load_analyzer.go and load_generator.go are both working as intended.

Execute with "go run scriptname.go"

Load_generator
 1. Prompts user for number or workloads. This is input an integer. Workloads include compute, network and storage, or anything if we understand how what we are observing works. 
 2. Prompts user for number of integers to assign to a given load.
 3. X number of workloads are created as individual json files.
 4. Each Json file contains a KVP workload names, number of loads (3 initial load lists) and the integers that compromise the individual loads, and a value random value. 

Load_analyzer
1. Ingests all correctly formatted json files with the name Workload*.json
2. Performs math (decipher it of your own accord, it's all open).
3. Returns the following for each workload.
    A. Workload name.
    B. Total load/cost of load1.
    C. Total load/cost of load2.
    D. Total load/cost of load3.
    E. Workload's load1 relative cost/load to all other workloads's load1 cost/load.
    F. Workload's load2 relative cost/load to all other workloads's load2 cost/load.
    G. Workload's load3 relative cost/load to all other workloads's load3 cost/load.
    H. The workload's total cost/load.
    I. The workload's relative total cost/load to other workloads's total cost/load.
    J. The workload's relative value to other workloads value. 

BenchMark Hardware: Ryzen 1920, 128GB 2666hz mem. 
12:54:00 Start 10000 workload 30000(3X loads with 10k floats) metrics generation. 
12:56:33 Finish 10000 workload 30000(3X loads with 10k floats) metric generation. 

12:57:00 Start workload analyzing. 
12:59:20 Finish workload analyzing.
100,000 datapoints. 

