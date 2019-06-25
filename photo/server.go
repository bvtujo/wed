// server.go

package main

import (
    "database/sql"
    "io"
    "net/http"
    "os"

    "github.com/labstack/echo"
    "github.com/labstack/echo/middleware"
    _ "github.com/mattn/go-sqlite3"
)

type Photo struct {
	ID int64 `json:"id"`
	Src string `json:"src"`
}

type PhotoCollection struct {
	Photos []Photo `json:"items"`
}

func initDb(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil || db == nil {
		panic("Error connecting to database")
	}

	return db
}

func migrateDb(db *sql.DB) {
	sql := `
		CREATE TABLE IF NOT EXISTS photos(
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			src VARCHAR NOT NULL,
			seqnum INTEGER NOT NULL
		);
	`

	_, err := db.Exec(sql)
	if err != nil {
		panic(err)
	}
}

func getPhotos(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT * FROM photos")
		if err != nil {
			panic(err)
		}

		defer rows.Close()

		result := PhotoCollection{}

		for rows.Next() {
			photo := Photo{}

			err2 := rows.Scan(&photo.ID, &photo.Src)
			if err2 != nil {
				panic(err2)
			}

			result.Photos = append(result.Photos, photo)
		}

		return c.JSON(http.StatusOK, result)
	}
}

func addPhotos(db *sql.DB) error {
	
	photo_path := "public/photos/"

	panic("Not implemented.")
}

func main() {
	db := initDb("database/db.sqlite")
	migrateDb(db)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.File("/", "public/index.html")
	e.GET("/photos", getPhotos(db))

	e.Logger.Fatal(e.Start(":9000"))
}