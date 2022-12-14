package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/go-sql-driver/mysql"
)

type Database struct {
	DB *sql.DB
}

func ConnectWithConnector(DB_USER *string, DB_PASS *string, DB_NAME *string, INSTANCE_CONNECTION_NAME *string) (*Database, error) {
	var database Database

	// Checking in env variables aren't properly loaded in
	if *DB_USER == "" || *DB_PASS == "" || *DB_NAME == "" || *INSTANCE_CONNECTION_NAME == "" {
		return nil, fmt.Errorf("Couldn't load env variables")
	}

	var (
		dbUser                 = DB_USER                  // e.g. 'my-db-user'
		dbPwd                  = DB_PASS                  // e.g. 'my-db-password'
		dbName                 = DB_NAME                  // e.g. 'my-database'
		instanceConnectionName = INSTANCE_CONNECTION_NAME // e.g.  maybe /cloudql/ 'project:region:instance'
		usePrivate             = os.Getenv("PRIVATE_IP")
	)

	d, err := cloudsqlconn.NewDialer(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cloudsqlconn.NewDialer: %v", err)
	}
	var opts []cloudsqlconn.DialOption
	// cloud sql is setup on public IP this will never be hit unless this is changed
	if usePrivate != "" {
		opts = append(opts, cloudsqlconn.WithPrivateIP())
	}

	// Connecting to cloud sql database
	mysql.RegisterDialContext("cloudsqlconn",
		func(ctx context.Context, addr string) (net.Conn, error) {
			return d.Dial(ctx, *instanceConnectionName, opts...)
		})

	dbURI := fmt.Sprintf("%s:%s@cloudsqlconn(localhost:3306)/%s?parseTime=true",
		*dbUser, *dbPwd, *dbName)

	// Opening mysql database
	database.DB, err = sql.Open("mysql", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}
	return &database, nil
}

type User struct {
	UUID     int64
	Username string
	Password string
}

func (d *Database) GetUser() ([]User, error) {
	rows, err := d.DB.Query("SELECT * FROM User")
	if err != nil {
		return nil, fmt.Errorf("error running query: %w", err)
		//gctx.Status(http.StatusInternalServerError)
		//return
	}
	users := []User{}
	defer rows.Close()
	for rows.Next() {
		var u User
		rows.Scan(&u.UUID, &u.Username, &u.Password)
		users = append(users, u)
	}
	rows.Err()

	return users, nil

}
