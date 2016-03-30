package ip2location

import (
    "encoding/binary"
    "os"
    "net"
)

func ipToInt(ip *net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

func getOffset(what [25]uint32, mid, baseaddr, dbcolumn, dbtype uint32, isIpv6 bool) uint32 {
    if isIpv6 {
        return baseaddr + mid * (dbcolumn * 4 + 12) + 12 + 4 * (what[dbtype]-1)
    }
    return baseaddr + mid * (dbcolumn * 4 + 0) + 0 + 4 * (what[dbtype]-1)
}

type ip2header struct {
    DatabaseType uint8
    DatabaseColumn uint8
    DatabaseYear uint8
    DatabaseMonth uint8
    DatabaseDay uint8
    IPv4Count uint32
    IPv4Addr uint32
    IPv6Count uint32
    IPv6Addr uint32
}

type IP2Location struct {

    // Unexported fields below
    hdr ip2header
    fd *os.File

    dbType uint32
    dbCol uint32
}

// IP2LocationEntry contains the results about a given IPv4/IPv6 address
type IP2LocationEntry struct {
    IP string
    CountryShort string
    CountryLong string
    Region string
    City string
    ISP string
    Latitude float32
    Longitude float32
    Domain string
    ZipCode string
    TimeZone string

    //XXX(cschmitt): These seem of little use to us?
    /*
    NetSpeed string
    IddCode string
    AreaCode string
    WeatherStationCode string
    WeatherStationName string
    MccPosition string
    MncPosition string
    MobileBroadband string
    Elevation string
    */

    UsageType string
}

// NewIP2Location grabs a new database context with a given db file
func NewIP2Location(dbpath string) (db IP2Location, err error) {
    var hdr ip2header

    fd, err := os.Open(dbpath)
    if err != nil {
        return
    }
    
    if err = binary.Read(fd, binary.LittleEndian, &hdr); err != nil {
        return
    }

    db.fd = fd
    db.hdr = hdr
    db.dbType = uint32(hdr.DatabaseType)
    db.dbCol = uint32(hdr.DatabaseColumn)
    
    return
}

func (s *IP2Location) read32(pos uint32) (ret uint32) {
    s.fd.Seek(int64(pos)-1, 0)
    binary.Read(s.fd, binary.LittleEndian, &ret)
    return 
}

func (s *IP2Location) reads(pos uint32) string {
    var sz uint8
    s.fd.Seek(int64(pos)-1, 0)
    binary.Read(s.fd, binary.LittleEndian, &sz)
    buf := make([]byte, sz)
    s.fd.Read(buf)
    return string(buf)
}

func (s *IP2Location) readfloat(pos uint32) (ret float32) {
    s.fd.Seek(int64(pos)-1, 0)
    binary.Read(s.fd, binary.LittleEndian, &ret)
    return 
}


func (s *IP2Location) readRecord(mid uint32, ip *net.IP) (rec IP2LocationEntry) {
    var (
        dbtype = s.hdr.DatabaseType
        baseaddr uint32
        isIPV6 = false
    )
    
    if ip.To4() != nil {
        baseaddr = s.hdr.IPv4Addr
    } else {
        baseaddr = s.hdr.IPv6Addr
        isIPV6 = true
    }
    
    if countryPosition[dbtype] != 0 {
        offz := s.read32(getOffset(countryPosition, mid, baseaddr, s.dbCol, s.dbType, isIPV6))
        rec.CountryShort = s.reads(offz + 1)
        rec.CountryLong = s.reads(offz + 4)
    }
    
    if regionPosition[dbtype] != 0 {
        rec.Region = s.reads(s.read32(getOffset(regionPosition, mid, baseaddr, s.dbCol, s.dbType, isIPV6)) + 1)
    }
    
    if cityPosition[dbtype] != 0 {
        rec.City = s.reads(s.read32(getOffset(cityPosition, mid, baseaddr, s.dbCol, s.dbType, isIPV6)) + 1)
    }
    
    if ispPosition[dbtype] != 0 {
        rec.ISP = s.reads(s.read32(getOffset(ispPosition, mid, baseaddr, s.dbCol, s.dbType, isIPV6)) + 1)
    }
    
    if latitudePosition[dbtype] != 0 {
        rec.Latitude = s.readfloat(getOffset(latitudePosition, mid, baseaddr, s.dbCol, s.dbType, isIPV6))
    }
    
     if longitudePosition[dbtype] != 0 {
        rec.Longitude = s.readfloat(getOffset(longitudePosition, mid, baseaddr, s.dbCol, s.dbType, isIPV6))
    }
    
    if domainPosition[dbtype] != 0 {
        rec.Domain = s.reads(s.read32(getOffset(domainPosition, mid, baseaddr, s.dbCol, s.dbType, isIPV6)) + 1)
    }
    
    if zipcodePosition[dbtype] != 0 {
        rec.ZipCode = s.reads(s.read32(getOffset(zipcodePosition, mid, baseaddr, s.dbCol, s.dbType, isIPV6)) + 1)
    }
    
    if timezonePosition[dbtype] != 0 {
        rec.TimeZone = s.reads(s.read32(getOffset(timezonePosition, mid, baseaddr, s.dbCol, s.dbType, isIPV6)) + 1)
    }
    
    if usageTypePosition[dbtype] != 0 {
        rec.UsageType = s.reads(s.read32(getOffset(usageTypePosition, mid, baseaddr, s.dbCol, s.dbType, isIPV6)) + 1)
    }

    return
}

// GetRecord returns an IP2LocationEntry if found within the database
func (s *IP2Location) GetRecord(ipstr string) (rec IP2LocationEntry, err error) {
    var (
        low, mid uint32
        high  = s.hdr.IPv4Count
        baseaddr = s.hdr.IPv4Addr
    )

    ip := net.ParseIP(ipstr)
    if ip == nil {
        err = errorInvalidIP
        return
    }
    /*
    if ip.To16() != nil {
        fmt.Println("wattt")
        high = s.hdr.IPv6Count
        baseaddr = s.hdr.IPv6Addr
    }
    */
    
    ipno := ipToInt(&ip)
    for low <= high {
        mid = ((low + high) / 2)
        ipfrom := s.read32(baseaddr + mid * s.dbCol * 4)
        ipto := s.read32(baseaddr + (mid + 1) * s.dbCol * 4)

        if ((ipno >= ipfrom) && (ipno < ipto)) {
            rec = s.readRecord(mid, &ip)
            return
        }

        if (ipno < ipfrom) {
            high = mid - 1
        } else {
            low = mid + 1
        }
    }
    err = errorNoRecordFound
    return
}

// Close the ip2location database
func (s *IP2Location) Close() {
    s.fd.Close()
}