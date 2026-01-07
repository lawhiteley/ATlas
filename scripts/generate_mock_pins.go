package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "../atlas_data.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS stored_pins (
        did TEXT PRIMARY KEY,
        uri TEXT,
        longitude REAL,
        latitude REAL,
        name TEXT,
        handle TEXT,
        description TEXT,
        website TEXT,
        avatar TEXT
    );`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("%q: %s\n", err, createTableSQL)
	}

	for i := 1; i <= 50; i++ {
		did := fmt.Sprintf("did:plc:%s", uuid.New().String())
		uri := fmt.Sprintf("at://%s/self", did)
		latitude := rand.Float64()*180 - 90
		longitude := rand.Float64()*360 - 180
		name := fmt.Sprintf("Name %d", i)
		handle := fmt.Sprintf("name.%d.io", i)
		description := fmt.Sprintf("Sample description %d.", i)
		website := fmt.Sprintf("https://%s", handle)
		avatar := "https://avatar.iran.liara.run/public?diff=" + strconv.Itoa(i)

		insertSQL := `
        INSERT INTO stored_pins (did, uri, longitude, latitude, name, handle, description, website, avatar)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);`

		if _, err := db.Exec(insertSQL, did, uri, longitude, latitude, name, handle, description, website, avatar); err != nil {
			log.Fatalf("%q: %s\n", err, insertSQL)
		}
	}
}
