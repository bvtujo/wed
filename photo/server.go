// server.go

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/mattn/go-sqlite3"
)

type Photo struct {
	ID  int64  `json:"id"`
	Src string `json:"src"`
}

type PhotoCollection struct {
	Photos []Photo `json:"items"`
}

func initDb(filepath string) *sql.DB {

	db, err := sql.Open("sqlite3", filepath)
	if err != nil || db == nil {
		fmt.Println("Error connecting to database")
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
		fmt.Println(err)
		panic(err)
	}
}

func getPhotos(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT * FROM photos")
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		defer rows.Close()

		result := PhotoCollection{}

		for rows.Next() {
			photo := Photo{}
			var seq int
			err2 := rows.Scan(&photo.ID, &photo.Src, &seq)
			if err2 != nil {
				panic(err2)
			}

			result.Photos = append(result.Photos, photo)
		}
		fmt.Println(result)
		return c.JSON(http.StatusOK, result)

	}
}

func addToDb(db *sql.DB, file os.FileInfo, basePath string) error {

	filePath := basePath + file.Name()

	var seq int
	_, err := fmt.Sscanf(filePath, "./photos/20190615_HenryEly_web-%d.jpeg", &seq)

	fleSrc := "127.0.0.1:8080/photos/" + file.Name()

	stmt, err := db.Prepare("INSERT INTO photos (src, seqnum) VALUES (?, ?)")

	defer stmt.Close()

	_, err = stmt.Exec(fileSrc, seq)
	fmt.Println("File " + filePath + " successfully added to database.")
	return err
}

func addPhotos(db *sql.DB) error {

	photo_path := "./public/photos/"
	db_path := "/public/photos/"
	f, err := os.Open(photo_path)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	files, err := f.Readdir(-1)
	f.Close()

	for _, file := range files {
		addToDb(db, file, db_path)
	}
	return err
}

func main() {
	db := initDb("database/db.sqlite")
	migrateDb(db)

	err := addPhotos(db)
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.File("/", "public/index.html")
	e.GET("/photos", getPhotos(db))

	e.Logger.Fatal(e.Start(":8080"))
}
