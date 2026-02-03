package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
)

func main() {
	for {
		if err := run(); err != nil {
			log.Fatal(err)
		}
		var again bool
		huh.NewConfirm().Title("Again?").Value(&again).Run()

		if !again {
			break
		}
	}
}

func run() error {
	var zip string
	huh.NewInput().Title("Enter zip code").Value(&zip).Run()
	lat, long := lookupZip(zip)

	var stores []Store
	spinner.New().Title("finding nearby stores...").
		Action(func() { stores = fetchStores(lat, long) }).Run()

	if len(stores) == 0 {
		fmt.Println("No stores found.")
		return nil
	}

	var selectedStore string
	options := make([]huh.Option[string], len(stores))
	for i, store := range stores {
		options[i] = huh.NewOption(store.Name, store.StoreNumber)
	}

	huh.NewSelect[string]().Title("Select a store").Options(options...).
		Height(5).Value(&selectedStore).Run()

	var menu []MenuItem
	spinner.New().Title("finding menu...").
		Action(func() { menu = fetchMenu(selectedStore) }).Run()

	var order string
	huh.NewInput().Title("Welcome to Taco Bell, can I take your order?").Value(&order).Run()

	result := parseOrder(order, menu)
	fmt.Println(formatOrder(result))

	return nil
}

func lookupZip(zip string) (float64, float64) {
	if len(zip) < 5 {
		return 0, 0
	}

	url := fmt.Sprintf("https://niiknow.github.io/zipcode-us/db/%s/%s.json", zip[:2], zip[:5])
	resp, err := http.Get(url)

	if err != nil {
		return 0, 0
	}
	defer resp.Body.Close()

	var data struct {
		Lat  float64 `json:"lat"`
		Long float64 `json:"lng"`
	}

	if jsonErr := json.NewDecoder(resp.Body).Decode(&data); jsonErr != nil {
		return 0, 0
	}

	return data.Lat, data.Long
}
