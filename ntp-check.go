package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

// NTP Packet structure
type ntpPacket struct {
	LI      uint8 // Leap Indicator
	VN      uint8 // Version Number
	Mode    uint8 // Mode
	Stratum uint8 // Stratum
	Poll    uint8 // Poll Interval
	Prec    int8  // Precision

	RootDelay         uint32 // Root Delay
	RootDispersion    uint32 // Root Dispersion
	ReferenceID       uint32 // Reference Identifier
	ReferenceTimeSec  uint32 // Reference Timestamp Seconds
	ReferenceTimeFrac uint32 // Reference Timestamp Fractional Seconds

	OriginateTimeSec  uint32 // Originate Timestamp Seconds
	OriginateTimeFrac uint32 // Originate Timestamp Fractional Seconds

	ReceiveTimeSec  uint32 // Receive Timestamp Seconds
	ReceiveTimeFrac uint32 // Receive Timestamp Fractional Seconds

	TransmitTimeSec  uint32 // Transmit Timestamp Seconds
	TransmitTimeFrac uint32 // Transmit Timestamp Fractional Seconds
}

func main() {
	ntpServer := "time.google.com:123" // Replace with your NTP server
	timeout := 5 * time.Second

	// 1. Create the NTP packet.
	packet := createNTPPacket()

	// 2. Connect to the NTP server.
	conn, err := net.Dial("udp", ntpServer)
	if err != nil {
		log.Fatalf("Failed to connect to NTP server: %v", err)
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		log.Fatalf("Failed to set deadline: %v", err)
	}

	// 3. Send NTP request
	packetData, err := packet.Marshal()
	if err != nil {
		log.Fatalf("Failed to marshal NTP Packet: %v", err)
	}

	_, err = conn.Write(packetData)
	if err != nil {
		log.Fatalf("Failed to send NTP request: %v", err)
	}

	// 4. Receive response.
	response := make([]byte, 48)
	n, err := conn.Read(response)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}
	if n != 48 {
		log.Fatalf("Invalid NTP response length: %d", n)
	}

	// 5. Unmarshal Response

	responsePacket := ntpPacket{}
	err = responsePacket.Unmarshal(response)
	if err != nil {
		log.Fatalf("Failed to unmarshal NTP response: %v", err)
	}

	// 6. print response info
	fmt.Println("NTP Server:", ntpServer)
	fmt.Println("Leap Indicator:", responsePacket.LI)
	fmt.Println("Version Number:", responsePacket.VN)
	fmt.Println("Mode:", responsePacket.Mode)
	fmt.Println("Stratum:", responsePacket.Stratum)
	fmt.Println("Root Delay (ms):", float64(responsePacket.RootDelay)/float64(1<<16)*1000)
	fmt.Println("Root Dispersion (ms):", float64(responsePacket.RootDispersion)/float64(1<<16)*1000)
	fmt.Println("Reference Time:", responsePacket.ReferenceTime())
	fmt.Println("Transmit Time:", responsePacket.TransmitTime())

	fmt.Println("Successfully received NTP response from server.")
}

// Marshal converts the NTP packet to a byte slice
func (p *ntpPacket) Marshal() ([]byte, error) {
	buf := make([]byte, 48)

	// First byte is LI, VN, and Mode combined
	buf[0] = (p.LI << 6) | (p.VN << 3) | (p.Mode)
	buf[1] = p.Stratum
	buf[2] = p.Poll
	buf[3] = byte(p.Prec) // Cast Prec to a byte

	binary.BigEndian.PutUint32(buf[4:8], p.RootDelay)
	binary.BigEndian.PutUint32(buf[8:12], p.RootDispersion)
	binary.BigEndian.PutUint32(buf[12:16], p.ReferenceID)
	binary.BigEndian.PutUint32(buf[16:20], p.ReferenceTimeSec)
	binary.BigEndian.PutUint32(buf[20:24], p.ReferenceTimeFrac)
	binary.BigEndian.PutUint32(buf[24:28], p.OriginateTimeSec)
	binary.BigEndian.PutUint32(buf[28:32], p.OriginateTimeFrac)
	binary.BigEndian.PutUint32(buf[32:36], p.ReceiveTimeSec)
	binary.BigEndian.PutUint32(buf[36:40], p.ReceiveTimeFrac)
	binary.BigEndian.PutUint32(buf[40:44], p.TransmitTimeSec)
	binary.BigEndian.PutUint32(buf[44:48], p.TransmitTimeFrac)

	return buf, nil
}

// Unmarshal fills an NTP packet from byte slice
func (p *ntpPacket) Unmarshal(data []byte) error {
	if len(data) != 48 {
		return fmt.Errorf("invalid NTP packet size: %d", len(data))
	}

	p.LI = data[0] >> 6
	p.VN = (data[0] >> 3) & 0x07
	p.Mode = data[0] & 0x07
	p.Stratum = data[1]
	p.Poll = data[2]
	p.Prec = int8(data[3])

	p.RootDelay = binary.BigEndian.Uint32(data[4:8])
	p.RootDispersion = binary.BigEndian.Uint32(data[8:12])
	p.ReferenceID = binary.BigEndian.Uint32(data[12:16])
	p.ReferenceTimeSec = binary.BigEndian.Uint32(data[16:20])
	p.ReferenceTimeFrac = binary.BigEndian.Uint32(data[20:24])
	p.OriginateTimeSec = binary.BigEndian.Uint32(data[24:28])
	p.OriginateTimeFrac = binary.BigEndian.Uint32(data[28:32])
	p.ReceiveTimeSec = binary.BigEndian.Uint32(data[32:36])
	p.ReceiveTimeFrac = binary.BigEndian.Uint32(data[36:40])
	p.TransmitTimeSec = binary.BigEndian.Uint32(data[40:44])
	p.TransmitTimeFrac = binary.BigEndian.Uint32(data[44:48])

	return nil
}

// ReferenceTime returns the reference time as a time.Time
func (p *ntpPacket) ReferenceTime() time.Time {
	return ntpTimestampToTime(p.ReferenceTimeSec, p.ReferenceTimeFrac)
}

// TransmitTime returns the transmit time as a time.Time
func (p *ntpPacket) TransmitTime() time.Time {
	return ntpTimestampToTime(p.TransmitTimeSec, p.TransmitTimeFrac)
}

// ntpTimestampToTime converts NTP time to go's time format
func ntpTimestampToTime(seconds uint32, fractions uint32) time.Time {
	const ntpEpochOffset = 2208988800
	const ntpFractionalScale = 4294967296

	sec := int64(seconds) - ntpEpochOffset

	nsec := int64(float64(fractions) / float64(ntpFractionalScale) * 1e9)

	return time.Unix(sec, nsec)
}

// createNTPPacket generates a NTPv4 packet with only the minimum info
func createNTPPacket() ntpPacket {
	return ntpPacket{
		LI:              0,
		VN:              4,
		Mode:            3,
		Stratum:         0,
		Poll:            0,
		Prec:            -6,                                     // Precision of the clock
		TransmitTimeSec: uint32(time.Now().Unix() + 2208988800), // Add epoch offset
	}
}
