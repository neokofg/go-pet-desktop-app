// cmd/app/main.go
package main

import (
	"context"
	"github.com/neokofg/go-pet-desktop-app/internal/handler"
	"github.com/neokofg/go-pet-desktop-app/internal/store"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// getAppDataPath возвращает путь к директории данных приложения
func getAppDataPath() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}

	appDir := filepath.Join(appData, "PasswordManager")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return "", err
	}

	return appDir, nil
}

func main() {
	// Получаем путь к директории данных
	appDir, err := getAppDataPath()
	if err != nil {
		return
	}

	// Настраиваем логирование
	logFile, err := os.OpenFile(filepath.Join(appDir, "app.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Получаем путь к исполняемому файлу для загрузки шаблонов
	executable, err := os.Executable()
	if err != nil {
		log.Printf("Failed to get executable path: %v\n", err)
		return
	}
	exeDir := filepath.Dir(executable)

	log.Printf("Application starting... App dir: %s, Exe dir: %s\n", appDir, exeDir)

	// Инициализируем хранилище
	store, err := store.NewSQLiteStore("passwords.db")
	if err != nil {
		log.Printf("Failed to initialize database: %v\n", err)
		return
	}
	defer store.Close()

	// Настраиваем Gin
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = logFile
	r := gin.Default()

	// Создаем handler
	h := handler.NewHandler(store)

	// Проверяем шаблоны
	templatesPath := filepath.Join(exeDir, "web", "templates")
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		log.Printf("Templates directory not found in: %s\n", templatesPath)
		return
	}

	r.LoadHTMLGlob(filepath.Join(templatesPath, "*"))
	r.Static("/static", filepath.Join(exeDir, "web", "static"))

	// Регистрируем маршруты
	r.GET("/", h.HandleIndex)
	r.POST("/credentials", h.HandleCreate)
	r.GET("/credentials", h.HandleList)
	r.GET("/credentials/search", h.HandleSearch)
	r.DELETE("/credentials/:id", h.HandleDelete)
	r.GET("/open-url", h.HandleOpenURL)

	// Запускаем приложение Wails
	err = wails.Run(&options.App{
		Title:  "Менеджер паролей",
		Width:  800,
		Height: 600,
		AssetServer: &assetserver.Options{
			Handler: r,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		OnStartup: func(ctx context.Context) {
			log.Println("Application started")
		},
		OnShutdown: func(ctx context.Context) {
			log.Println("Application closing...")
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
