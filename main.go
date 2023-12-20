package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Flexible Octopus (22nd June 2023 - 30th June 2023)

// Flexible Octopus (1st July 2023 - 21st July 2023)
// 29.62
// 57.31

// Flexible Octopus (22nd July 2023 - 21st August 2023)
// 29.62
// 57.31

// Flexible Octopus (22nd August 2023 - 21st September 2023)
// 29.62
// 57.31

// Flexible Octopus (22nd September 2023 - 30th September 2023)
// 29.62
// 57.31

// Flexible Octopus (1st October 2023 - 21st October 2023)
// 26.91
// 57.31

// Flexible Octopus (22nd October 2023 - 21st November 2023)
// 26.91 p / kWh
// 57.31 p / day

var firstPeriodStarts = time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)
var lastPeriodStarts = time.Date(2023, 10, 30, 23, 30, 0, 0, time.UTC)

func main() {
	usage, err := loadUsageCSV()
	if err != nil {
		panic(err)
	}

	for _, record := range usage {
		fmt.Printf("ending %s : %.4f kWh\n", record.HalfHourEndingAt, record.Usage)
		break
	}

	prices, err := loadAgilePrices()
	if err != nil {
		panic(err)
	}

	for _, record := range prices {
		fmt.Printf("ending %s : %.4f p / kWh\n", record.PeriodTo, record.AgileImportPrice)
		break
	}

	missing := 0
	cumulativeCost := 0.0
	cumulativeConsumption := 0.0

	for start := firstPeriodStarts.UTC(); start.Before(lastPeriodStarts); start = start.Add(30 * time.Minute) {
		end := start.Add(30 * time.Minute)

		consumption, ok := usage[mapKey(end)]
		if !ok {
			panic(fmt.Errorf("no usage for %s", end))
		}

		price, ok := prices[mapKey(end)]
		if !ok {
			fmt.Printf("⚠️ missing price for %s", end)
			missing++
			continue
		}

		fmt.Printf(
			"%s—%s used %.4f kWh @ %.2f p\n",
			start.Format("2006-01-02 15:04"),
			end.Format("15:04"),
			consumption.Usage,
			price.AgileImportPrice,
		)

		cumulativeConsumption += consumption.Usage
		cumulativeCost += consumption.Usage * price.AgileImportPrice
	}

	if missing > 0 {
		fmt.Printf("⚠️ %d missing periods excluded\n", missing)
	}

	fmt.Printf("\n     Period: %s to %s\nConsumption: %.2f kWh\n Agile cost: £%.2f\n",
		firstPeriodStarts.Format("02 January 2006 at 15:04"),
		lastPeriodStarts.Format("02 January 2006 at 15:04"),
		cumulativeConsumption,
		cumulativeCost/100,
	)

}

func mapKey(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

// Usage represents a single row in the CSV.
type Usage struct {
	HalfHourEndingAt time.Time
	Usage            float64
}

// AgilePrice represents a single row in the CSV.
type AgilePrice struct {
	PeriodFrom       time.Time
	PeriodTo         time.Time
	AgileImportPrice float64
	AgileExportPrice float64
}

func loadUsageCSV() (map[string]Usage, error) {
	// Open the CSV file.
	file, err := os.Open("data/electricity_records.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a new CSV reader.
	reader := csv.NewReader(file)

	// Read and ignore the header.
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	records := make(map[string]Usage)
	for {
		row, err := reader.Read()
		if err != nil {
			if err == csv.ErrFieldCount {
				continue // handle wrong number of fields
			}
			break // stop at EOF or other errors
		}

		// Parse the half-hour ending timestamp.
		timestamp, err := time.Parse("2006-01-02 15:04:05-07:00", row[0])
		if err != nil {
			return nil, fmt.Errorf("error parsing timestamp: %v", err)
		}

		// Parse the usage value.
		usage, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing usage: %v", err)
		}

		records[mapKey(timestamp)] = Usage{
			HalfHourEndingAt: timestamp,
			Usage:            usage,
		}
	}

	return records, nil
}

// loadAgilePrices takes a file path and returns a slice of PeriodRecords and an error, if any.
func loadAgilePrices() (map[string]AgilePrice, error) {
	file, err := os.Open("data/agile-half-hour-actual-rates-01-11-2022_20-12-2023.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a new CSV reader.
	reader := csv.NewReader(file)

	// Read and ignore the header.
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	records := make(map[string]AgilePrice)
	for {
		row, err := reader.Read()
		if err != nil {
			if err == csv.ErrFieldCount {
				continue // handle wrong number of fields
			}
			break // stop at EOF or other errors
		}

		// Parse the period start and end times.
		periodFrom, err := time.Parse("02/01/2006 15:04", row[0])
		if err != nil {
			fmt.Printf("Error parsing Period From: %v\n", err)
			continue
		}

		periodTo, err := time.Parse("02/01/2006 15:04", row[1])
		if err != nil {
			fmt.Printf("Error parsing Period To: %v\n", err)
			continue
		}

		// Parse the import and export prices.
		agileImportPrice, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			fmt.Printf("Error parsing Agile Import Price: %v\n", err)
			continue
		}

		agileExportPrice, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			fmt.Printf("Error parsing Agile Export Price: %v\n", err)
			continue
		}

		records[mapKey(periodTo)] = AgilePrice{
			PeriodFrom:       periodFrom,
			PeriodTo:         periodTo,
			AgileImportPrice: agileImportPrice,
			AgileExportPrice: agileExportPrice,
		}
	}

	return records, nil
}
