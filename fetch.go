package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type Store struct {
	Name        string
	StoreNumber string
}

type MenuItem struct {
	Name  string
	Price float64
}

func fetchStores(lat, long float64) []Store {
	url := fmt.Sprintf("https://www.tacobell.com/tacobellwebservices/v4/tacobell/stores?latitude=%f&longitude=%f", lat, long)
	resp, err := http.Get(url)

	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var data struct {
		NearByStores []struct {
			StoreNumber string `json:"storeNumber"`
			Address     struct {
				Line1  string `json:"line1"`
				Town   string `json:"town"`
				Region struct {
					Isocode string `json:"isocode"`
				} `json:"region"`
			} `json:"address"`
		} `json:"nearByStores"`
	}

	if storeJsonErr := json.NewDecoder(resp.Body).Decode(&data); storeJsonErr != nil {
		return nil
	}

	stores := make([]Store, 0, 7)
	for i, s := range data.NearByStores {
		if i >= 7 {
			break
		}

		state := strings.TrimPrefix(s.Address.Region.Isocode, "US-")
		stores = append(stores, Store{
			Name:        fmt.Sprintf("%s %s, %s", s.Address.Line1, s.Address.Town, state),
			StoreNumber: s.StoreNumber,
		})
	}
	return stores
}

func fetchMenu(storeNumber string) []MenuItem {
	url := fmt.Sprintf("https://www.tacobell.com/tacobellwebservices/v4/tacobell/products/menu/%s", storeNumber)
	resp, err := http.Get(url)

	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var data struct {
		MenuProductCategories []struct {
			Products []struct {
				Name  string `json:"name"`
				Price struct {
					Value float64 `json:"value"`
				} `json:"price"`
			} `json:"products"`
		} `json:"menuProductCategories"`
	}
	if menuJsonErr := json.NewDecoder(resp.Body).Decode(&data); menuJsonErr != nil {
		return nil
	}

	var items []MenuItem
	symbolsRegex := regexp.MustCompile(`[®™©℠]`)

	for _, category := range data.MenuProductCategories {
		for _, product := range category.Products {
			if product.Price.Value > 0 {
				name := symbolsRegex.ReplaceAllString(product.Name, "")
				items = append(items, MenuItem{
					Name:  strings.ToLower(name),
					Price: product.Price.Value,
				})
			}
		}
	}
	return items
}
