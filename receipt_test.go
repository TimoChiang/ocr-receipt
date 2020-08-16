package main

import (
	"bufio"
	"os"
	"reflect"
	"testing"
)

func TestReceipt_ScanStore(t *testing.T) {
	r := new(Receipt)
	if r.Store != nil {
		t.Fatal("initial error: Store is not nil")
	}

	for i, test := range []struct {
		keyword string
		store Store
	}{
		{
			"オーケー",
			&Ok{},
		},
	} {
		store := r.scanStore(test.keyword)
		if !reflect.DeepEqual(store, test.store) {
			t.Errorf("#%d: expected %v, got %v", i, test.store, store)
		}
	}
}

func TestOk_ScanData(t *testing.T) {
	store := &Ok{}
	for _, test := range []struct {
		inputFile string
		expect []*Product
	}{
		{
			"test/ok.txt",
			[]*Product{
				{"キッコーマンエンブンヒカエメショウユ", 1,255,0}, // normal
				{"バナメイムキエヒ", 1,279,8}, // Discount 1
				{"ヤマザキ ショウジュン8マイ", 1,78,3}, // Discount 2
				{"コンドウキュウニュウ1000ml", 2,318,0}, // Quantity more then 1
				{"コクサンワカトリムネニク", 1,149,0}, // scan Price word "半"
				{"ニチレイブロッコリー250g", 1,152,0}, // scan Price word "ギ"
				{"メキシコサンブタモモキリオトシ", 1,401,120}, // last Price Discount
			},
		},
	} {
		file, err := os.Open(test.inputFile)
		if err != nil {
			t.Fatalf("test file open error: %v", err)
		}
		scanner := bufio.NewScanner(file)

		products, err := store.ScanData(scanner)

		if err != nil {
			t.Errorf("function ScanData got error: %v", err)
		}
		for i, p := range products {
			if !reflect.DeepEqual(p, test.expect[i]) {
				t.Errorf("Product#%d: expected %v, got %v", i, test.expect[i], p)
			}
		}
		file.Close()
	}
}

