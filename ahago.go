//ahago uses the AVM Home Automation HTTP Interface to let you control
//your AVM Home Automation Products (06.06.2014)
package ahago

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

type connection struct {
	sid      string
	time     time.Time
	pass     string
	username string
}

//Connect fetches a new Session-ID with the provided username and password
//It returns a connection struct
func Connect(username string, password string) *connection {
	connection_new := new(connection)
	connection_new.username = username
	connection_new.pass = password
	connection_new.getSessionId()
	connection_new.time = time.Now()
	return connection_new
}

//GetStatus gives information about all AVM Home Automation devices in your system
func (c *connection) GetStatus() {
	switchListString := c.GetSwitchList("")
	if switchListString == "" {
		fmt.Println("No Devices.")
	}
	switchList := strings.Split(switchListString, ",")
	for _, val := range switchList {
		val = strings.TrimSpace(val)
		name := strings.TrimSpace(c.GetSwitchName(val))
		fmt.Printf("Name: %v     AIN: %v\n", name, val)
		present := strings.TrimSpace(c.GetSwitchPresent(val))
		present_int, _ := strconv.ParseInt(present, 10, 64)
		if present_int == 1 {
			state := strings.TrimSpace(c.GetSwitchState(val))
			state_int, _ := strconv.ParseInt(state, 10, 32)
			if state_int == 1 {
				fmt.Println(name, " is on.")
				power := strings.TrimSpace(c.GetSwitchPower(val))
				fmt.Printf("Current Power: %vmW\n", power)
			} else {
				fmt.Println(name, " is off.")
			}
		} else {
			fmt.Println(name, " is not connected.")
		}
		energy := strings.TrimSpace(c.GetSwitchEnergy(val))
		fmt.Printf("Total Energy used: %vWh\n", energy)
		fmt.Println()
	}
}

//GetSwitchName gets a device identifier
//and returns the name of the device as a string.
func (s *connection) GetSwitchName(ain string) string {
	result := s.prepareSH(ain, "getswitchname")
	return result
}

//GetSwitchEnergy gets a device identifier
//and returns the used Energy in Wh as a string.
func (s *connection) GetSwitchEnergy(ain string) string {
	result := s.prepareSH(ain, "getswitchenergy")
	return result
}

//GetSwitchPower gets a device identifier
//and returns the current Power in mW as a string.
func (s *connection) GetSwitchPower(ain string) string {
	result := s.prepareSH(ain, "getswitchpower")
	return result
}

//GetSwitchPresent gets a device identifier
//and returns wether the device is connected over DECT (1) or not(0)  as a string.
func (s *connection) GetSwitchPresent(ain string) string {
	result := s.prepareSH(ain, "getswitchpresent")
	return result
}

//GetSwitchState gets a device identifier
//and returns wether the device is on(1), off(0) or unknown(inval) a string.
func (s *connection) GetSwitchState(ain string) string {
	result := s.prepareSH(ain, "getswitchstate")
	return result
}

//SetSwitchToggle gets a device identifier, switches the device to the opposite state
//and returns the new state(1/0) of the device as a string.
func (s *connection) SetSwitchToggle(ain string) string {
	result := s.prepareSH(ain, "setswitchtoggle")
	return result
}

//SetSwitchOff gets a device identifier. switches the device off
//and returns the new state(0) as a string.
func (s *connection) SetSwitchOff(ain string) string {
	result := s.prepareSH(ain, "setswitchoff")
	return result
}

//SetSwitchOn gets a device identifier. switches the device on
//and returns the new state(1) as a string.
func (s *connection) SetSwitchOn(ain string) string {
	res := s.prepareSH(ain, "setswitchon")
	return res
}

//GetSwitchList gets a device identifier
//and returns a list of all devices in the system as a string.
func (s *connection) GetSwitchList(ain string) string {
	res := s.prepareSH(ain, "getswitchlist")
	return res
}

func (c *connection) prepareSH(ain, switchcmd string) string {
	baseurl := "http://fritz.box/webservices/homeautoswitch.lua"
	parameters := make(map[string]string)
	if time.Since(c.time).Minutes() >= 10 {
		c.getSessionId()
	}
	c.time = time.Now()
	parameters["sid"] = c.sid
	if ain != "" {
		parameters["ain"] = ain
	}
	parameters["switchcmd"] = switchcmd
	Url := prepareRequest(baseurl, parameters)
	return string(sendRequest(Url))
}

func prepareRequest(baseUrl string, parameters_in map[string]string) *url.URL {
	var Url *url.URL
	Url, err := url.Parse(baseUrl)
	if err != nil {
		panic("Fehler beim URL parsen: ")
	}
	parameters := url.Values{}
	for key, value := range parameters_in {
		parameters.Add(key, value)
	}
	Url.RawQuery = parameters.Encode()
	//fmt.Println("HTTP GET: ", Url.String())
	return Url
}

//sendRequest sends a HTTP Request based on the given URL
//and returns the answer as a bytearray
func sendRequest(Url *url.URL) []byte {
	resp, err := http.Get(Url.String())
	if err != nil {
		fmt.Println("Error:", err)
		panic("Fehler beim Request senden")
	}
	if resp.StatusCode != 200 {
		fmt.Println("Response: ", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("Fehler beim lesen des response Body")
	}
	//time.Sleep(100 * time.Millisecond)
	return body
}

//getSessionID fetches a session-id based on the username and password in the connection struct
func (c *connection) getSessionId() {
	loginUrl := "http://fritz.box/login_sid.lua"
	parameters := make(map[string]string)
	body := sendRequest(prepareRequest(loginUrl, parameters))
	type Result struct {
		SID       string
		Challenge string
		BlockTime string
	}
	v := Result{SID: "none", Challenge: "none", BlockTime: "none"}
	err := xml.Unmarshal(body, &v)
	if err != nil {
		panic("Fehler bei Unmarshalling")
	}
	//fmt.Printf("SID: %q\n", v.SID)
	//fmt.Printf("Challenge: %q\n", v.Challenge)
	//fmt.Printf("Blocktime in s: %q\n", v.BlockTime)
	if v.SID == "0000000000000000" {
		parameters["username"] = c.username
		parameters["response"] = getSIDResponse(v.Challenge, c.pass)
		body = sendRequest(prepareRequest(loginUrl, parameters))
		err := xml.Unmarshal(body, &v)
		if err != nil {
			panic("Fehler bei Unmarshalling2")
		}
		//fmt.Printf("SID: %q\n", v.SID)
		//fmt.Printf("Challenge: %q\n", v.Challenge)
		//fmt.Printf("Blocktime in s: %q\n", v.BlockTime)
	}
	c.sid = v.SID
}

//getSIDResponse gets the Fritzbox challenge and the password
//It returns the response
func getSIDResponse(challenge string, pass string) string {
	hash := md5.New()
	utf8 := []byte(challenge + "-" + pass)
	utf16le := utf8ToUtf16le(utf8)
	hash.Write(utf16le)
	md5hash := hex.EncodeToString(hash.Sum(nil))
	return challenge + "-" + md5hash
}

//utf8ToUtf16le gets an UTF8 bytearray
//and returns an UTF16LE bytearray
func utf8ToUtf16le(in []byte) []byte {
	var ucps []rune
	var utf16uint []uint16
	var utf16le []byte
	for len(in) > 0 {
		r, size := utf8.DecodeRune(in)
		ucps = append(ucps, r)
		in = in[size:]
	}
	utf16uint = utf16.Encode(ucps)
	b := make([]byte, 2)
	for val, _ := range utf16uint {
		binary.LittleEndian.PutUint16(b, utf16uint[val])
		utf16le = append(utf16le, b[0], b[1])
	}
	return utf16le
}

//Close invalidates the Session-Id.
func (c *connection) Close() {
	baseurl := "http://fritz.box/webservices/homeautoswitch.lua"
	parameters := make(map[string]string)
	parameters["sid"] = c.sid
	parameters["logout"] = "logout"
	Url := prepareRequest(baseurl, parameters)
	sendRequest(Url)
}
