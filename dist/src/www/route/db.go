package route

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var (
	SALT = ""
	DB   *sql.DB
)

// init sqlite3 database
func InitDB(salt string, secretKey string, adminPwd string) error {
	SALT = salt

	db, err := sql.Open("sqlite3", "./casino.db")
	if err != nil {
		return err
	}
	DB = db
    
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uname TEXT UNIQUE NOT NULL,
		pwd TEXT NOT NULL
	);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf(`INSERT INTO users(uname, pwd) VALUES('admin', '%s');`,
		sha1Hash([]byte(adminPwd+SALT))))
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS secret(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		secret TEXT UNIQUE NOT NULL
	);`)
	if err != nil {
		return err
	}

	// backup my secret key into DB
	_, err = db.Exec(fmt.Sprintf(`INSERT INTO secret(secret) VALUES('%s');`, secretKey))
	if err != nil {
		return err
	}

	return nil
}
