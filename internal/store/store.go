package store

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/neokofg/go-pet-desktop-app/internal/model"
	"log"
	"os"
	"path/filepath"
)

type Store interface {
	Create(cred *model.Credential) error
	List() ([]model.Credential, error)
	Delete(id int) error
	Close() error
}

type SQLiteStore struct {
	db *sql.DB
}

func getAppDataPath() (string, error) {
	// Получаем путь к AppData в Windows
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// Если APPDATA не определена, используем домашнюю директорию
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}

	// Создаем директорию для нашего приложения
	appDir := filepath.Join(appData, "PasswordManager")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return "", err
	}

	return appDir, nil
}

func NewSQLiteStore(dbName string) (*SQLiteStore, error) {
	appDir, err := getAppDataPath()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(appDir, dbName)
	log.Printf("Opening database at: %s\n", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Failed to open database: %v\n", err)
		return nil, err
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		log.Printf("Failed to ping database: %v\n", err)
		db.Close()
		return nil, err
	}

	// Изменяем структуру таблицы - убираем NOT NULL для username
	createTable := `
    CREATE TABLE IF NOT EXISTS credentials (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        service TEXT NOT NULL,
        username TEXT,
        password TEXT NOT NULL,
        notes TEXT
    );`

	if _, err := db.Exec(createTable); err != nil {
		log.Printf("Failed to create table: %v\n", err)
		db.Close()
		return nil, err
	}

	log.Println("Database initialized successfully")
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Create(cred *model.Credential) error {
	log.Printf("Creating credential for service: %s\n", cred.Service)

	result, err := s.db.Exec(
		"INSERT INTO credentials (service, username, password, notes) VALUES (?, ?, ?, ?)",
		cred.Service, cred.Username, cred.Password, cred.Notes,
	)
	if err != nil {
		log.Printf("Failed to insert credential: %v\n", err)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v\n", err)
		return err
	}

	cred.ID = int(id)
	log.Printf("Created credential with ID: %d\n", id)
	return nil
}

func (s *SQLiteStore) List() ([]model.Credential, error) {
	log.Println("Listing all credentials")

	rows, err := s.db.Query("SELECT id, service, username, password, notes FROM credentials")
	if err != nil {
		log.Printf("Failed to query credentials: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var credentials []model.Credential
	for rows.Next() {
		var cred model.Credential
		if err := rows.Scan(&cred.ID, &cred.Service, &cred.Username, &cred.Password, &cred.Notes); err != nil {
			log.Printf("Failed to scan credential row: %v\n", err)
			return nil, err
		}
		credentials = append(credentials, cred)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during row iteration: %v\n", err)
		return nil, err
	}

	log.Printf("Found %d credentials\n", len(credentials))
	return credentials, nil
}

func (s *SQLiteStore) Delete(id int) error {
	log.Printf("Deleting credential with ID: %d\n", id)

	result, err := s.db.Exec("DELETE FROM credentials WHERE id = ?", id)
	if err != nil {
		log.Printf("Failed to delete credential: %v\n", err)
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get affected rows: %v\n", err)
		return err
	}

	log.Printf("Deleted %d rows\n", affected)
	return nil
}

func (s *SQLiteStore) Close() error {
	log.Println("Closing database connection")
	return s.db.Close()
}
