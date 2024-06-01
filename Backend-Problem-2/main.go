package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ProductName  string  `json:"productName"`
	Price        int     `json:"price"`
	Rating       float64 `json:"rating"`
	Discount     int     `json:"discount"`
	Availability string  `json:"availability"`
	Company      string  `json:"company,omitempty"`
	ID           string  `json:"id"`
}

type ProductResponse struct {
	Products []Product `json:"products"`
}

var (
	client      = &http.Client{Timeout: 10 * time.Second}
	serverURL   = "http://20.244.56.144/test/companies"
	companies   = []string{"AMZ", "FLP", "SNP", "MYN", "AZO"}
	productData = make(map[string]Product)
	dataMutex   sync.Mutex
)

func fetchProducts(company, category string, minPrice, maxPrice, top int) ([]Product, error) {
	url := fmt.Sprintf("%s/%s/categories/%s/products?top=%d&minPrice=%d&maxPrice=%d", serverURL, company, category, top, minPrice, maxPrice)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var products []Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, err
	}

	return products, nil
}

func getProductsHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	category := pathParts[2]

	top, _ := strconv.Atoi(r.URL.Query().Get("top"))
	if top == 0 {
		top = 10
	}
	minPrice, _ := strconv.Atoi(r.URL.Query().Get("minPrice"))
	maxPrice, _ := strconv.Atoi(r.URL.Query().Get("maxPrice"))
	sortBy := r.URL.Query().Get("sortBy")
	order := r.URL.Query().Get("order")

	var allProducts []Product
	var wg sync.WaitGroup
	for _, company := range companies {
		wg.Add(1)
		go func(company string) {
			defer wg.Done()
			products, err := fetchProducts(company, category, minPrice, maxPrice, top)
			if err != nil {
				log.Printf("Error fetching products for company %s: %v", company, err)
				return
			}

			dataMutex.Lock()
			for _, product := range products {
				product.ID = uuid.New().String()
				product.Company = company
				allProducts = append(allProducts, product)
				productData[product.ID] = product
			}
			dataMutex.Unlock()
		}(company)
	}
	wg.Wait()

	sort.Slice(allProducts, func(i, j int) bool {
		switch sortBy {
		case "price":
			if order == "desc" {
				return allProducts[i].Price > allProducts[j].Price
			}
			return allProducts[i].Price < allProducts[j].Price
		case "rating":
			if order == "desc" {
				return allProducts[i].Rating > allProducts[j].Rating
			}
			return allProducts[i].Rating < allProducts[j].Rating
		case "discount":
			if order == "desc" {
				return allProducts[i].Discount > allProducts[j].Discount
			}
			return allProducts[i].Discount < allProducts[j].Discount
		case "company":
			if order == "desc" {
				return allProducts[i].Company > allProducts[j].Company
			}
			return allProducts[i].Company < allProducts[j].Company
		default:
			return false
		}
	})

	n := top
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	start := (page - 1) * n
	end := start + n
	if end > len(allProducts) {
		end = len(allProducts)
	}
	if start > len(allProducts) {
		start = len(allProducts)
	}
	paginatedProducts := allProducts[start:end]

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ProductResponse{Products: paginatedProducts}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func getProductHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	productID := pathParts[4]

	dataMutex.Lock()
	product, exists := productData[productID]
	dataMutex.Unlock()

	if !exists {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/categories/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/categories/") && strings.Count(r.URL.Path, "/") == 3 {
			getProductsHandler(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/categories/") && strings.Count(r.URL.Path, "/") == 4 {
			getProductHandler(w, r)
		} else {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
		}
	})

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
