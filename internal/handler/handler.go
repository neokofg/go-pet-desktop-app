// internal/handler/handler.go
package handler

import (
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/neokofg/go-pet-desktop-app/internal/model"
	"github.com/neokofg/go-pet-desktop-app/internal/store"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store store.Store
}

func NewHandler(store store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) HandleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func (h *Handler) HandleCreate(c *gin.Context) {
	log.Println("Handling create request")
	var cred model.Credential

	cred.Service = c.PostForm("service")
	cred.Username = c.PostForm("username")
	cred.Password = c.PostForm("password")
	cred.Notes = c.PostForm("notes")

	log.Printf("Received credential: service=%s, username=%s\n", cred.Service, cred.Username)

	// Изменяем валидацию - проверяем только service и password
	if cred.Service == "" || cred.Password == "" {
		log.Println("Validation failed: missing required fields")
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Service and password are required",
		})
		return
	}

	if err := h.store.Create(&cred); err != nil {
		log.Printf("Failed to create credential: %v\n", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to save credential",
		})
		return
	}

	log.Printf("Created credential with ID: %d\n", cred.ID)
	c.HTML(http.StatusOK, "credential-row.html", cred)
}

func (h *Handler) HandleList(c *gin.Context) {
	credentials, err := h.store.List()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load credentials",
		})
		return
	}

	c.HTML(http.StatusOK, "credentials-list.html", gin.H{
		"credentials": credentials,
	})
}

func (h *Handler) HandleDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid ID",
		})
		return
	}

	if err := h.store.Delete(id); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to delete credential",
		})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) HandleSearch(c *gin.Context) {
	query := strings.ToLower(c.Query("query"))
	log.Printf("Search query: %s\n", query)

	// Если запрос пустой, возвращаем весь список
	if query == "" {
		h.HandleList(c)
		return
	}

	credentials, err := h.store.List()
	if err != nil {
		log.Printf("Failed to load credentials for search: %v\n", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to search credentials",
		})
		return
	}

	// Фильтруем записи
	var filtered []model.Credential
	for _, cred := range credentials {
		if strings.Contains(strings.ToLower(cred.Service), query) ||
			strings.Contains(strings.ToLower(cred.Username), query) ||
			strings.Contains(strings.ToLower(cred.Notes), query) {
			filtered = append(filtered, cred)
		}
	}

	log.Printf("Found %d matching credentials\n", len(filtered))

	// Если ничего не найдено, показываем специальный шаблон
	if len(filtered) == 0 {
		c.HTML(http.StatusOK, "credentials-list.html", gin.H{
			"noResults": true,
			"query":     query,
		})
		return
	}

	c.HTML(http.StatusOK, "credentials-list.html", gin.H{
		"credentials": filtered,
		"query":       query,
	})
}

// HandleOpenURL обрабатывает открытие ссылок в системном браузере
func (h *Handler) HandleOpenURL(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// Добавляем протокол, если его нет
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	log.Printf("Opening URL in browser: %s\n", url)

	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("cmd", "/c", "start", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default: // linux и другие
		err = exec.Command("xdg-open", url).Start()
	}

	if err != nil {
		log.Printf("Failed to open URL: %v\n", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
