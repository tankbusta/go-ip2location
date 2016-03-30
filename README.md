# Go ip2location

A Golang client for reading data from [ip2location](https://github.com/tankbusta/go-ip2location) IPv4 and IPv6 databases.

## Usage

    package main
    
    import (
        "github.com/tankbusta/go-ip2location"
        "log"
        "path/filepath"
    )
    
    func main() {
        tp := filepath.Join("./", "testdata", "IP-COUNTRY.BIN")
        db, err := NewIP2Location(tp)
        if err != nil {
            log.Fatalf("Failed to open database: %v\n", err)
        }
        defer db.Close()

        rec, err :=  db.GetRecord("19.5.10.1")
        if err != nil {
            log.Fatalf("Failed to get record: %v\n", err)
        }
        log.Println(rec)
    }

## Installation

    $ go get github.com/tankbusta/go-ip2location
   
## Testing

    $ go test -v

## Future Plans

    * Allow for loading the entire database into memory
