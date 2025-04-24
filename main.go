package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Produk struct {
	ID         int    `json:"id"`
	Nama       string `json:"nama"`
	Harga      int    `json:"harga"`
	KategoriID int    `json:"kategori_id"`
	Kategori   string `json:"kategori"`
}

type Kategori struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

var DB *sql.DB

func main() {
	var err error
	DB, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/uts_web")
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", Home).Methods("GET")
	router.HandleFunc("/produk/store", StoreProduk).Methods("POST")
	router.HandleFunc("/produk/update/{id}", UpdateProduk).Methods("POST")
	router.HandleFunc("/produk/delete/{id}", DeleteProduk).Methods("POST")
	router.HandleFunc("/produk/data/{id}", GetProdukData).Methods("GET")

	http.Handle("/", router)
	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Home(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	editID := r.URL.Query().Get("edit")
	var editProduk Produk
	if editID != "" {
		row := DB.QueryRow("SELECT id, nama, harga, kategori_id FROM produk WHERE id = ?", editID)
		row.Scan(&editProduk.ID, &editProduk.Nama, &editProduk.Harga, &editProduk.KategoriID)
	}

	rows, _ := DB.Query("SELECT id, nama FROM kategori")
	var kategoris []Kategori
	for rows.Next() {
		var k Kategori
		rows.Scan(&k.ID, &k.Nama)
		kategoris = append(kategoris, k)
	}

	rows2, _ := DB.Query(`SELECT produk.id, produk.nama, produk.harga, produk.kategori_id, kategori.nama FROM produk JOIN kategori ON produk.kategori_id = kategori.id`)
	var produks []Produk
	for rows2.Next() {
		var p Produk
		rows2.Scan(&p.ID, &p.Nama, &p.Harga, &p.KategoriID, &p.Kategori)
		produks = append(produks, p)
	}

	tmpl.Execute(w, map[string]interface{}{
		"Produk":    editProduk,
		"Kategoris": kategoris,
		"Produks":   produks,
	})
}

func StoreProduk(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	nama := r.FormValue("nama")
	harga, _ := strconv.Atoi(r.FormValue("harga"))
	kategoriID, _ := strconv.Atoi(r.FormValue("kategori_id"))
	DB.Exec("INSERT INTO produk (nama, harga, kategori_id) VALUES (?, ?, ?)", nama, harga, kategoriID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func UpdateProduk(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	params := mux.Vars(r)
	id := params["id"]
	nama := r.FormValue("nama")
	harga, _ := strconv.Atoi(r.FormValue("harga"))
	kategoriID, _ := strconv.Atoi(r.FormValue("kategori_id"))
	DB.Exec("UPDATE produk SET nama=?, harga=?, kategori_id=? WHERE id=?", nama, harga, kategoriID, id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func DeleteProduk(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	DB.Exec("DELETE FROM produk WHERE id=?", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func GetProdukData(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var produk Produk
	row := DB.QueryRow("SELECT id, nama, harga, kategori_id FROM produk WHERE id = ?", id)
	err := row.Scan(&produk.ID, &produk.Nama, &produk.Harga, &produk.KategoriID)
	if err != nil {
		http.Error(w, "Produk tidak ditemukan", http.StatusNotFound)
		return
	}

	// Kirim data produk dalam format JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produk)
}
