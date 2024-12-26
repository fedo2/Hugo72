package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// Config reprezentuje strukturu konfiguračního souboru.
// Obsahuje sekci Phase3, která definuje parametry pro připojení k FTP serveru
// a seznam souborů, které mají být nahrány.
//
// Struktura konfiguračního souboru (config.json):
// {
//   "phase3": {
//     "ftpHost": "ftp.example.com",
//     "ftpUser": "uzivatel",
//     "ftpPassword": "heslo",
//     "remoteDir": "/cesta/na/serveru",
//     "files_to_upload": ["soubor1.txt", "soubor2.txt"]
//   }
// }
//
// Pokud soubor chybí nebo je poškozen, program skončí s chybou a nevykoná žádnou akci.
type Config struct {
	Phase3 struct {
		FtpHost      string   `json:"ftpHost"`       // Adresa FTP serveru (např. "ftp.example.com")
		FtpUser      string   `json:"ftpUser"`       // Uživatelské jméno pro připojení k FTP
		FtpPassword  string   `json:"ftpPassword"`   // Heslo pro připojení k FTP
		RemoteDir    string   `json:"remoteDir"`     // Cílový adresář na FTP serveru, kam budou soubory nahrány
		FilesToUpload []string `json:"files_to_upload"` // Seznam lokálních souborů určených k nahrání na FTP server
	} `json:"phase3"`
}

// connectToFtp se připojí k FTP serveru pomocí zadaných přihlašovacích údajů.
// Vrací připojení k serveru nebo chybu, pokud se připojení nezdaří.
func connectToFtp(ftpServer, ftpUser, ftpPassword string) (*ftp.ServerConn, error) {
	// Pokus o připojení k FTP serveru s nastavením timeoutu 5 sekund.
	conn, err := ftp.Dial(ftpServer, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("chyba při připojování k FTP serveru: %w", err)
	}

	// Přihlášení na FTP server pomocí poskytnutých přihlašovacích údajů.
	if err := conn.Login(ftpUser, ftpPassword); err != nil {
		return nil, fmt.Errorf("chyba při přihlášení na FTP server: %w", err)
	}

	log.Println("Úspěšně připojeno k FTP serveru.")
	return conn, nil
}

// uploadFile nahraje jeden soubor na FTP server do zadaného adresáře.
// Používá stejné jméno souboru pro lokální i vzdálený soubor.
//
// Očekávaný formát souborů je takový, že každý soubor musí být dostupný
// v lokálním souborovém systému a jeho název odpovídá tomu, co je uvedeno
// v konfiguraci. Tato funkce otevře soubor podle jeho názvu, přejde
// do cílového adresáře na serveru a nahraje soubor pod stejným názvem.
func uploadFile(conn *ftp.ServerConn, remoteDir, localFile string) error {
	// Otevření lokálního souboru k nahrání.
	file, err := os.Open(localFile)
	if err != nil {
		return fmt.Errorf("chyba při otevření lokálního souboru '%s': %w", localFile, err)
	}
	defer file.Close()

	// Změna adresáře na FTP serveru na cílový adresář.
	if err := conn.ChangeDir(remoteDir); err != nil {
		return fmt.Errorf("chyba při změně adresáře na serveru '%s': %w", remoteDir, err)
	}

	// Nahrání souboru na server.
	if err := conn.Stor(localFile, file); err != nil {
		return fmt.Errorf("chyba při nahrávání souboru '%s' na server: %w", localFile, err)
	}

	log.Printf("Soubor '%s' byl úspěšně nahrán na server.\n", localFile)
	return nil
}

// loadConfig načte a dekóduje konfigurační soubor ze zadané cesty.
// Vrací strukturu Config nebo chybu při načítání či dekódování.
func loadConfig(filePath string) (*Config, error) {
	// Otevření konfiguračního souboru.
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("chyba při otevírání souboru konfigurace '%s': %w", filePath, err)
	}
	defer file.Close()

	// Dekódování obsahu souboru do struktury Config.
	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("chyba při dekódování konfigurace: %w", err)
	}
	return &config, nil
}

// main je vstupní bod programu.
// Načte konfiguraci, připojí se k FTP serveru a nahraje soubory zadané v konfiguraci.
func main() {
	// Načtení konfigurace z konfiguračního souboru.
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Chyba při načítání konfigurace: %v", err)
	}

	// Připojení k FTP serveru s využitím údajů z konfigurace.
	conn, err := connectToFtp(config.Phase3.FtpHost, config.Phase3.FtpUser, config.Phase3.FtpPassword)
	if err != nil {
		log.Fatalf("Chyba: %v", err)
	}
	defer conn.Quit()

	// Iterujeme přes seznam souborů, které mají být nahrány.
	// Každý soubor je nejprve očištěn od mezer na začátku a na konci názvu.
	// Pokud je jméno prázdné (například z neplatného záznamu), přeskočíme ho.
	for _, file := range config.Phase3.FilesToUpload {
		file = strings.TrimSpace(file) // Odstranění mezer okolo názvu souboru
		if file == "" {
			continue // Přeskočení prázdných položek v seznamu
		}

		// Pokus o nahrání každého souboru na FTP server
		if err := uploadFile(conn, config.Phase3.RemoteDir, file); err != nil {
			log.Printf("Chyba při nahrávání souboru '%s': %v\n", file, err)
		}
	}
}
