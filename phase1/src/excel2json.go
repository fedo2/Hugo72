package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

type Config struct {
	Phase1 struct {
		InputFile  string `json:"inputFile"`
		OutputFile string `json:"outputFile"`
	} `json:"phase1"`
}

type Data72 struct {
	Info  Info   `json:"info"`
	Users []User `json:"users"`
}

type Info struct {
	LastUpdate   string `json:"lastUpdate"`
	Nadpis       string `json:"nadpis"`
	Zprava       string `json:"zprava"`
	PocetZaznamu int64  `json:"pocetZaznamu"`
	PocetAno     int64  `json:"pocetAno"`
}

type User struct {
	Jmeno  string `json:"Jmeno"`
	Email  string `json:"email"`
	Prijde Prijde `json:"Prijde"`
}

type Prijde string

const (
	Ano   Prijde = "Ano"
	Empty Prijde = ""
	Ne    Prijde = "Ne"
)

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Println("Chyba při načítání konfigurace:", err)
		return
	}

	excelFile, err := openExcelFile(config.Phase1.InputFile)
	if err != nil {
		fmt.Println("Chyba při otevírání Excel souboru:", err)
		return
	}
	defer excelFile.Close()

	sheetName := excelFile.GetSheetName(0)
	rows, err := excelFile.GetRows(sheetName)
	if err != nil {
		fmt.Println("Chyba při čtení řádků ze souboru:", err)
		return
	}

	data := processRows(rows)
	err = writeJSONFile(config.Phase1.OutputFile, data)
	if err != nil {
		fmt.Println("Chyba při zápisu JSON souboru:", err)
		return
	}

	fmt.Println("Soubor", config.Phase1.OutputFile, "byl úspěšně vytvořen.")
}

func loadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func openExcelFile(filePath string) (*excelize.File, error) {
	return excelize.OpenFile(filePath)
}

func processRows(rows [][]string) Data72 {
	var data Data72

	if len(rows) > 1 {
		data.Info.Nadpis = rows[0][1]
		data.Info.Zprava = rows[1][1]
	}
	totalRecords := 0
	totalAno := 0

	for i, row := range rows {
		if i < 3 { // Přeskočení hlavičky
			continue
		}
		if len(row) < 6 || row[1] == "" { // Konec dat (prázdný řádek)
			break
		}

		user := User{
			Jmeno:  row[1],
			Email:  row[5],
			Prijde: Prijde(row[0]),
		}
		if user.Prijde == Ano {
			totalAno++
		}
		data.Users = append(data.Users, user)
		totalRecords++
	}

	data.Info.LastUpdate = time.Now().Format("2.1.2006 15.04.05")
	data.Info.PocetZaznamu = int64(totalRecords)
	data.Info.PocetAno = int64(totalAno)

	return data
}

func writeJSONFile(filePath string, data Data72) error {
	jsonFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ") // Pro lepší čitelnost JSON souboru
	return encoder.Encode(data)
}
