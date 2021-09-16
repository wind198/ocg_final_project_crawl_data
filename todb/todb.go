package todb

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

func ConnectDB() *sql.DB {
	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "ecommerce",
		AllowNativePasswords: true,
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	if pingErr := db.Ping(); pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println(db, *db)
	fmt.Println("Connected")
	return db
}

func InsertToDB(db *sql.DB, category string, name string, price float64, imageSource string) (int64, error) {
	q := `insert into newTable(productCategory,productName,productPrice,productImageSource)
	values(?,?,?,?)`
	res, err := db.Exec(q, category, name, price, imageSource)
	if err != nil {
		log.Printf("Error inserting data of product %q: %v", name, err)
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error inserting data of product %q: %v", name, err)
		return 0, err
	}
	return rows, nil
}
