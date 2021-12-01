package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
)

type Item struct {
	ID              uuid.UUID `json:"id"`
	CreatedAt       time.Time `json:"_"`
	UpdatedAt       time.Time `json:"_"`
	ProductName     string    `json:"productname"`
	Descript        string    `json:"descript"`
	ProductId       uuid.UUID `json:"product"`
	PriceInCents    int64     `json:"price_in_cents"`
	ProductLocation string    `json:"productlocation"`
}

func (i *Item) Create(conn *pgx.Conn, userID string) error {
	i.ProductName = strings.Trim(i.ProductName, " ")
	if len(i.ProductName) < 1 {
		return fmt.Errorf("ProductName must not be empty.")
	}
	if len(i.ProductLocation) < 1 {
		return fmt.Errorf("productlocation must not be empty.")
	}
	if i.PriceInCents < 0 {
		i.PriceInCents = 0
	}
	now := time.Now()

	row := conn.QueryRow(context.Background(), "INSERT INTO item (productname, descript, seller_id, price_in_cents, productlocation,created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6,$7) RETURNING id, seller_id", i.ProductName, i.Descript, userID, i.PriceInCents, i.ProductLocation, now, now)

	err := row.Scan(&i.ID, &i.ProductId)

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("There was an error creating the item")
	}

	return nil
}

func GetAllItems(conn *pgx.Conn) ([]Item, error) {
	rows, err := conn.Query(context.Background(), "SELECT id, productname, descript, seller_id, price_in_cents,productlocation, FROM item")
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("Error getting items")
	}

	var items []Item
	for rows.Next() {
		item := Item{}
		err = rows.Scan(&item.ID, &item.ProductName, &item.ProductLocation, &item.Descript, &item.ProductId, &item.PriceInCents)
		if err != nil {
			fmt.Println(err)
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

func GetItemsBeingSoldByUser(userID string, conn *pgx.Conn) ([]Item, error) {
	rows, err := conn.Query(context.Background(), "SELECT id, productname, productlocation, price_in_cents, descript, seller_id  FROM item WHERE seller_id = $1", userID)
	if err != nil {
		fmt.Printf("Error getting items %v", err)
		return nil, fmt.Errorf("There was an error getting the items")
	}

	var items []Item
	for rows.Next() {
		i := Item{}
		err = rows.Scan(&i.ID, &i.ProductName, &i.ProductLocation, &i.PriceInCents, &i.Descript, &i.ProductId)
		if err != nil {
			fmt.Printf("Error scaning item: %v", err)
			continue
		}
		items = append(items, i)
	}

	return items, nil
}

func (i *Item) Update(conn *pgx.Conn) error {
	i.ProductName = strings.Trim(i.ProductName, " ")
	if len(i.ProductName) < 1 {
		return fmt.Errorf("ProductName must not be empty")
	}
	if len(i.ProductLocation) < 1 {
		return fmt.Errorf("ProductLocation must not be empty")
	}

	// if i.ProductLocation < 0 {
	// 	i.ProductLocation = 0
	// }
	now := time.Now()
	_, err := conn.Exec(context.Background(), "UPDATE item SET productname=$1, productlocation=$6, descript=$2, price_in_cents=$3, updated_at=$4 WHERE id=$5", i.ProductName, i.ProductLocation, i.Descript, i.PriceInCents, now, i.ID)

	if err != nil {
		fmt.Printf("Error updating item: (%v)", err)
		return fmt.Errorf("Error updating item")
	}

	return nil
}

func FindItemById(id uuid.UUID, conn *pgx.Conn) (Item, error) {
	row := conn.QueryRow(context.Background(), "SELECT productname, descript, seller_id, price_in_cents FROM item WHERE id=$1", id)
	item := Item{
		ID: id,
	}
	err := row.Scan(&item.ProductName, &item.Descript, &item.ProductId, &item.PriceInCents)
	if err != nil {
		return item, fmt.Errorf("The item doesn't exist")
	}

	return item, nil
}
