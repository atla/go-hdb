/*
Copyright 2014 SAP SE

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	socks "github.com/armon/go-socks5"
)

// globals
var (
	// TestDSN (data source name for testing) has to be provided by calling go test with dsn parameter.
	TestDSN string
	// TestDropSchema could be provided by calling go test with dropSchema parameter.
	// If set to true (default), the test schema will be dropped after successful test execution.
	// If set to false, the test schema will remain on database after test execution.
	TestDropSchema bool
	// TestSchema will be used as test schema name and created on database by TestMain.
	// The scheam name consists of the prefix "test_" and a random Identifier.
	TestSchema Identifier
)

func TestMain(m *testing.M) {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	flag.StringVar(&TestDSN, "dsn", "hdb://user:password@ip_address:port", "database dsn")
	flag.BoolVar(&TestDropSchema, "dropSchema", true, "drop test schema after test ran successfully")

	if !flag.Parsed() {
		flag.Parse()
	}

	// run SOCKS server
	server, err := socks.New(&socks.Config{})
	listener, err := net.Listen("tcp", "127.0.0.1:1080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	go server.Serve(listener)

	// init driver
	db, err := sql.Open(DriverName, TestDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create schema
	TestSchema = RandomIdentifier("test_")
	if _, err := db.Exec(fmt.Sprintf("create schema %s", TestSchema)); err != nil {
		log.Fatal(err)
	}
	log.Printf("created schema %s", TestSchema)

	exitCode := m.Run()

	if exitCode == 0 && TestDropSchema {
		if _, err := db.Exec(fmt.Sprintf("drop schema %s cascade", TestSchema)); err != nil {
			log.Fatal(err)
		}
		log.Printf("dropped schema %s", TestSchema)
	}

	os.Exit(exitCode)
}
