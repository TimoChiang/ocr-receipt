package main

import (
	"bufio"
	vision "cloud.google.com/go/vision/apiv1"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Receipt struct {
	Store
}

type Product struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"`
	Discount int    `json:"discount"`
}

// scan from text file
func (r *Receipt) ScanFromTextFilePath(filePath string) ([]*Product, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// determine which store this receipt is from
	for scanner.Scan() {
		if store := r.scanStore(scanner.Text()); store != nil {
			r.Store = store
			break
		}
	}

	// ScanFromTextFilePath data
	if r.Store != nil {
		return r.ScanData(scanner)
	}
	return nil, errors.New("store not found")
}

// scan from image file
func (r *Receipt) ScanFromImageFile(filePath string) ([]*Product, error) {
	// access api to get receipt text, write into a txt file
	txtPath := ocrAndWriteFile(filePath)
	return r.ScanFromTextFilePath(txtPath)
}

// find the store by keywords
func(r *Receipt) scanStore(text string) (store Store) {
	if strings.Contains(text, "オーケー"){
		store = new(Ok)
	}
	return
}

// Store interface
type Store interface {
	ScanData(*bufio.Scanner) ([]*Product, error)
}

// Store : OK
type Ok struct {

}

func(s *Ok) ScanData(scanner *bufio.Scanner) ([]*Product, error) {
	startProductName := false
	startProductPrice := false
	productCount := 0
	numRe := regexp.MustCompile("[0-9]+")
	products := make([]*Product, 0)
	for scanner.Scan() {
		text := scanner.Text()
		// Next is start to get product name
		if strings.HasPrefix(text, "チNO") {
			startProductName = true
			continue
		}

		// Finish get product name
		if strings.HasPrefix(text, "割引前合計") || strings.HasPrefix(text, "F食料品3/103割引") {
			startProductName = false
			continue
		}

		// Next is start to get product Price
		if strings.HasPrefix(text, "※印") || strings.HasPrefix(text, "*印") {
			startProductPrice = true
			continue
		}

		// get product name (start with "F")
		if startProductName == true && strings.HasPrefix(text, "F") {
			product := &Product{Name:strings.TrimSpace(strings.TrimPrefix(text, "F")), Quantity:1}
			products = append(products, product)
		}

		// get product Quantity (start with "コX単")
		if startProductName == true && strings.Contains(text, "コX単") {
			numStrings := numRe.FindAllString(text, -1)
			if len(numStrings) > 0 {
				if quantity, err := strconv.Atoi(numStrings[0]); err == nil {
					products[len(products) - 1].Quantity = quantity
				} else {
					fmt.Println(numStrings[0], "is not an integer.")
				}
			} else {
				fmt.Println("can not find number in ", numStrings)
			}
		}

		// get product Price
		if startProductPrice == true {
			priceString := strings.Join(numRe.FindAllString(text, -1), "")
			if price, err := strconv.Atoi(priceString); err == nil {
				if strings.HasPrefix(text, "¥") || strings.HasPrefix(text, "半") || strings.HasPrefix(text, "ギ") {
					// finish
					if len(products) == productCount {
						break
					}
					products[productCount].Price = price
					productCount++
				} else {
					// Discount
					products[productCount - 1].Discount = price
				}
			} else {
				fmt.Println(priceString, "is not an integer.")
			}
		}
	}
	return products, nil
}

func ocrAndWriteFile(imagePath string) string {
	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Sets the name of the image file to annotate.
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	defer file.Close()

	image, err := vision.NewImageFromReader(file)
	if err != nil {
		log.Fatalf("Failed to create image: %v", err)
	}

	annotations, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		log.Fatalf("Failed to get OCR: %v", err)
	}

	// write content into txt file
	txtPath := "file/" + strconv.FormatInt(time.Now().Unix(), 10) + ".txt"
	f, err := os.Create(txtPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if len(annotations) == 0 {
		fmt.Println("No text found.")
	} else {
		fmt.Println("Write into file:")
		for _, annotation := range annotations {
			_, err := f.WriteString(annotation.Description + "\n")
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return txtPath
}