package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)
// scan receipt image and return json products information
func scan (w http.ResponseWriter, r *http.Request) {
	// get image from post
	file, header, err := r.FormFile("image")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// save image
	imagePath := "image/" + strconv.FormatInt(time.Now().Unix(), 10) + "_" + header.Filename
	out, err := os.Create(imagePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	ocr := New()
	products, err := ocr.ScanFromImageFile(imagePath)
	if err != nil {
		log.Fatalln(err)
	}
	js, err := json.Marshal(products)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// test json products information
func test (w http.ResponseWriter, r *http.Request) {
	fmt.Println("access!")
	ocr := New()
	file := "test/ok.txt"
	products, err := ocr.ScanFromTextFilePath(file)
	if err != nil {
		log.Fatalln(err)
	}

	js, err := json.Marshal(products)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func New() *Receipt {
	return new(Receipt)
}

func main() {
	http.HandleFunc("/api/v1/test", test)
	http.HandleFunc("/api/v1/scan", scan)
	http.ListenAndServe(":3000", nil)
}