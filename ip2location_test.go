package ip2location

import (
    "testing"
    "path/filepath"
    
    "github.com/stretchr/testify/assert"
)

func TestGetTrialRecord(t *testing.T) {
    tp := filepath.Join("./", "testdata", "IP-COUNTRY.BIN")
    db, err := NewIP2Location(tp)
    defer db.Close()

    assert.Nil(t, err, "NewIP2Location returned an error :(")

    rec, err :=  db.GetRecord("19.5.10.1")
    assert.Nil(t, err, "GetRecord returned an error :(")
    assert.Equal(t, rec.CountryShort, "US")
    assert.Equal(t, rec.CountryLong, "United States")
}

func BenchmarkGetRecord(b *testing.B) {
    tp := filepath.Join("./", "testdata", "IP-COUNTRY.BIN")
    db, err := NewIP2Location(tp)
    defer db.Close()
    b.ResetTimer()

    assert.Nil(b, err, "NewIP2Location returned an error :(")
    for n := 0; n < b.N; n++ {
        db.GetRecord("19.5.10.1")
    }
}