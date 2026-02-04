package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// --- STRUKTUR DATA ---

type Article struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Date    string `json:"date"`
	Content string `json:"content"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type LoginResponse struct {
	Token string `json:"token"`
}

// --- GLOBAL VARIABLES ---
var (
	articles  = []Article{}
	mu        sync.Mutex                    // Mutex untuk thread-safety
	jwtKey    = []byte("rahasia_super_aman") // Ganti dengan key yang kuat di production
	dataFile  = "data.json"                 // Nama file penyimpanan
)

func main() {
	// 1. Load data dari JSON saat server start
	loadData()

	mux := http.NewServeMux()

	// Public Routes
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/articles", handleArticles)       // GET (Public), POST (Admin)
	mux.HandleFunc("/articles/", handleArticleDetail) // GET (Public), PUT/DELETE (Admin)

	// Wrap dengan CORS Middleware
	handler := enableCORS(mux)

	fmt.Println("Server berjalan di http://localhost:8080")
	fmt.Println("Penyimpanan data menggunakan file:", dataFile)
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// --- FILE PERSISTENCE HELPERS ---

// loadData membaca data.json ke variabel memory 'articles'
func loadData() {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Open(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Jika file belum ada, buat seed data awal
			articles = []Article{
				{
					ID:      1723000000001,
					Title:   "Selamat Datang di Blog (File JSON)",
					Date:    "August 12, 2024",
					Content: "Artikel ini disimpan dalam file data.json. Server restart tidak akan menghapusnya.",
				},
			}
			saveDataInternal() // Simpan seed data ke file
			fmt.Println("File data.json baru dibuat dengan data awal.")
			return
		}
		fmt.Println("Gagal membuka file data:", err)
		return
	}
	defer file.Close()

	// Decode JSON ke slice articles
	bytes, _ := io.ReadAll(file)
	json.Unmarshal(bytes, &articles)
	fmt.Println("Berhasil memuat", len(articles), "artikel dari data.json")
}

// saveData menyimpan state 'articles' saat ini ke data.json
// Fungsi ini thread-safe (mengunci mutex sendiri)
func saveData() {
	mu.Lock()
	defer mu.Unlock()
	saveDataInternal()
}

// saveDataInternal adalah logika simpan tanpa lock (karena dipanggil oleh fungsi yang sudah nge-lock)
func saveDataInternal() {
	file, err := os.Create(dataFile) // Create akan menimpa/truncate file lama
	if err != nil {
		fmt.Println("Error saat menyimpan data:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Agar file JSON mudah dibaca manusia (pretty print)
	if err := encoder.Encode(articles); err != nil {
		fmt.Println("Error encoding JSON:", err)
	}
}

// --- HANDLERS ---

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Hardcode Admin Check
	if creds.Username == "admin" && creds.Password == "admin123" {
		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			Username: creds.Username,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(LoginResponse{Token: tokenString})
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

func handleArticles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET: List Articles (Public)
	if r.Method == http.MethodGet {
		// Kita perlu lock saat membaca memori agar tidak race condition dengan writer
		mu.Lock()
		json.NewEncoder(w).Encode(articles)
		mu.Unlock()
		return
	}

	// POST: Create Article (Protected)
	if r.Method == http.MethodPost {
		if !isAuthorized(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var newArticle Article
		if err := json.NewDecoder(r.Body).Decode(&newArticle); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Update Memory & File
		mu.Lock()
		newArticle.ID = time.Now().UnixMilli()
		articles = append([]Article{newArticle}, articles...) // Prepend
		saveDataInternal() // Simpan ke JSON
		mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newArticle)
		return
	}
}

func handleArticleDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	idStr := strings.TrimPrefix(r.URL.Path, "/articles/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// GET: Detail (Public)
	if r.Method == http.MethodGet {
		article, found := findArticleByID(id)
		if !found {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(article)
		return
	}

	// Protected Routes (Update/Delete)
	if !isAuthorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// PUT: Update
	if r.Method == http.MethodPut {
		var updatedData Article
		if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()
		for i, a := range articles {
			if a.ID == id {
				articles[i].Title = updatedData.Title
				articles[i].Date = updatedData.Date
				articles[i].Content = updatedData.Content
				
				saveDataInternal() // Simpan ke JSON

				json.NewEncoder(w).Encode(articles[i])
				return
			}
		}
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// DELETE: Delete
	if r.Method == http.MethodDelete {
		mu.Lock()
		defer mu.Unlock()
		for i, a := range articles {
			if a.ID == id {
				articles = append(articles[:i], articles[i+1:]...)
				
				saveDataInternal() // Simpan ke JSON
				
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "Deleted"}`))
				return
			}
		}
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
}

// --- HELPER FUNCTIONS ---

func findArticleByID(id int64) (Article, bool) {
	mu.Lock()
	defer mu.Unlock()
	for _, a := range articles {
		if a.ID == id {
			return a, true
		}
	}
	return Article{}, false
}

func isAuthorized(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}
	
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return false
	}
	return true
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}