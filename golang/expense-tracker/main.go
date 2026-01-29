package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// --- Models ---

type Expense struct {
	ID          int       `json:"id"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
}

type Config struct {
	Expenses []Expense `json:"expenses"`
	NextID   int       `json:"next_id"`
	Budget   float64   `json:"budget"` // Anggaran bulanan
}

const fileName = "expenses.json"

// --- Storage Logic ---

func loadData() Config {
	file, err := os.ReadFile(fileName)
	if err != nil {
		// Jika file tidak ada, kembalikan konfigurasi default
		return Config{Expenses: []Expense{}, NextID: 1, Budget: 0}
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{Expenses: []Expense{}, NextID: 1, Budget: 0}
	}
	return config
}

func saveData(config Config) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Gagal memproses data: %v\n", err)
		return
	}
	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		fmt.Printf("Gagal menyimpan ke file: %v\n", err)
	}
}

// --- Features ---

func addExpense(desc string, amount float64, category string) {
	if amount <= 0 {
		fmt.Println("Error: Jumlah (amount) harus bernilai positif.")
		return
	}

	config := loadData()
	newExpense := Expense{
		ID:          config.NextID,
		Date:        time.Now(),
		Description: desc,
		Amount:      amount,
		Category:    category,
	}

	config.Expenses = append(config.Expenses, newExpense)
	config.NextID++

	// Cek Anggaran Bulanan
	currentMonth := time.Now().Month()
	currentYear := time.Now().Year()
	var monthlyTotal float64
	for _, e := range config.Expenses {
		if e.Date.Month() == currentMonth && e.Date.Year() == currentYear {
			monthlyTotal += e.Amount
		}
	}

	saveData(config)
	fmt.Printf("Pengeluaran berhasil ditambahkan (ID: %d)\n", newExpense.ID)

	if config.Budget > 0 && monthlyTotal > config.Budget {
		fmt.Printf("⚠️ PERINGATAN: Anda telah melebihi anggaran bulanan sebesar $%.2f! (Terpakai: $%.2f)\n", config.Budget, monthlyTotal)
	}
}

func updateExpense(id int, desc string, amount float64, category string) {
	config := loadData()
	found := false

	for i, e := range config.Expenses {
		if e.ID == id {
			if desc != "" {
				config.Expenses[i].Description = desc
			}
			if amount > 0 {
				config.Expenses[i].Amount = amount
			}
			if category != "" {
				config.Expenses[i].Category = category
			}
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Error: Pengeluaran dengan ID %d tidak ditemukan.\n", id)
		return
	}

	saveData(config)
	fmt.Printf("Pengeluaran ID %d berhasil diperbarui.\n", id)
}

func listExpenses(categoryFilter string) {
	config := loadData()
	if len(config.Expenses) == 0 {
		fmt.Println("Belum ada data pengeluaran.")
		return
	}

	fmt.Printf("%-5s %-12s %-20s %-10s %-10s\n", "ID", "Tanggal", "Deskripsi", "Jumlah", "Kategori")
	fmt.Println(strings.Repeat("-", 65))

	for _, e := range config.Expenses {
		if categoryFilter != "" && !strings.EqualFold(e.Category, categoryFilter) {
			continue
		}
		fmt.Printf("%-5d %-12s %-20s $%-10.2f %-10s\n",
			e.ID, e.Date.Format("2006-01-02"), e.Description, e.Amount, e.Category)
	}
}

func deleteExpense(id int) {
	config := loadData()
	index := -1
	for i, e := range config.Expenses {
		if e.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		fmt.Printf("Error: Pengeluaran dengan ID %d tidak ditemukan.\n", id)
		return
	}

	config.Expenses = append(config.Expenses[:index], config.Expenses[index+1:]...)
	saveData(config)
	fmt.Println("Pengeluaran berhasil dihapus.")
}

func showSummary(month int) {
	config := loadData()
	var total float64
	year := time.Now().Year()

	if month > 0 {
		if month < 1 || month > 12 {
			fmt.Println("Error: Bulan tidak valid (1-12).")
			return
		}
		for _, e := range config.Expenses {
			if int(e.Date.Month()) == month && e.Date.Year() == year {
				total += e.Amount
			}
		}
		fmt.Printf("Total pengeluaran untuk %s: $%.2f\n", time.Month(month).String(), total)
	} else {
		for _, e := range config.Expenses {
			total += e.Amount
		}
		fmt.Printf("Total seluruh pengeluaran: $%.2f\n", total)
	}
}

func exportCSV() {
	config := loadData()
	if len(config.Expenses) == 0 {
		fmt.Println("Tidak ada data untuk diekspor.")
		return
	}

	csvFile, err := os.Create("expenses_export.csv")
	if err != nil {
		fmt.Printf("Gagal membuat file CSV: %v\n", err)
		return
	}
	defer csvFile.Close()

	fmt.Fprintln(csvFile, "ID,Date,Description,Amount,Category")
	for _, e := range config.Expenses {
		fmt.Fprintf(csvFile, "%d,%s,%s,%.2f,%s\n",
			e.ID, e.Date.Format("2006-01-02"), e.Description, e.Amount, e.Category)
	}
	fmt.Println("Data berhasil diekspor ke file expenses_export.csv")
}

// --- Main CLI Handler ---

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Gunakan: expense-tracker [command] [options]")
		fmt.Println("Perintah tersedia: add, list, delete, update, summary, budget, export")
		return
	}

	command := os.Args[1]

	switch command {
	case "add":
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		desc := addCmd.String("description", "", "Deskripsi pengeluaran")
		amount := addCmd.Float64("amount", 0, "Jumlah pengeluaran")
		category := addCmd.String("category", "General", "Kategori pengeluaran")
		addCmd.Parse(os.Args[2:])

		if *desc == "" || *amount <= 0 {
			fmt.Println("Error: description dan amount (positif) wajib diisi.")
			return
		}
		addExpense(*desc, *amount, *category)

	case "update":
		updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
		id := updateCmd.Int("id", 0, "ID pengeluaran yang akan diubah")
		desc := updateCmd.String("description", "", "Deskripsi baru")
		amount := updateCmd.Float64("amount", 0, "Jumlah baru")
		category := updateCmd.String("category", "", "Kategori baru")
		updateCmd.Parse(os.Args[2:])

		if *id == 0 {
			fmt.Println("Error: ID wajib diisi.")
			return
		}
		updateExpense(*id, *desc, *amount, *category)

	case "list":
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)
		cat := listCmd.String("category", "", "Filter berdasarkan kategori")
		listCmd.Parse(os.Args[2:])
		listExpenses(*cat)

	case "delete":
		deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
		id := deleteCmd.Int("id", 0, "ID pengeluaran yang akan dihapus")
		deleteCmd.Parse(os.Args[2:])
		if *id == 0 {
			fmt.Println("Error: ID wajib diisi.")
			return
		}
		deleteExpense(*id)

	case "summary":
		summaryCmd := flag.NewFlagSet("summary", flag.ExitOnError)
		month := summaryCmd.Int("month", 0, "Bulan spesifik (1-12)")
		summaryCmd.Parse(os.Args[2:])
		showSummary(*month)

	case "budget":
		budgetCmd := flag.NewFlagSet("budget", flag.ExitOnError)
		amount := budgetCmd.Float64("amount", 0, "Atur anggaran bulanan")
		budgetCmd.Parse(os.Args[2:])
		config := loadData()
		config.Budget = *amount
		saveData(config)
		fmt.Printf("Anggaran bulanan diatur sebesar $%.2f\n", *amount)

	case "export":
		exportCSV()

	default:
		fmt.Printf("Perintah tidak dikenal: %s\n", command)
	}
}
