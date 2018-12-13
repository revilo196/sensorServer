package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sensorServer/secure"
	"sensorServer/sensordata"
	"sensorServer/writemail"
	"strconv"
	"time"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

var lastTimes []time.Time

func panicLogger() {

	if r := recover(); r != nil {
		Error.Printf("<h2>PANIC:</h2> %v", r)
		panic(r)
	}

}

func putHandler(w http.ResponseWriter, r *http.Request) {
	defer panicLogger()

	if r.Method == "PUT" {

		//#READ INPUT
		b := make([]byte, r.ContentLength)
		n, err := r.Body.Read(b)
		if err != nil && err != io.EOF {
			Error.Println("error:", err, n)
			Error.Println(b)
			return
		}

		//#DECRYPT
		plaintext := secure.Decrypt(b)

		//CHECK HASH
		valid := secure.CheckHash(plaintext)
		if !valid {
			_, err = fmt.Fprintf(w, "HASH")
			if err != nil {
				Error.Println("error:", err)
			}
			return
		}

		//CHECK-IDENT
		u := plaintext[:15]
		valid = secure.CheckID(u)
		if !valid {
			_, err = fmt.Fprintf(w, "DENIED")
			if err != nil {
				Error.Println("error:", err)
			}
			return
		}

		_, err = fmt.Fprintf(w, "OK")
		if err != nil {
			Error.Println("error:", err)
		}

		sensornum := plaintext[15]
		pack := plaintext[:len(plaintext)-16]
		values := pack[16:]

		// U: IDENTIFIER  |  sensornum: Nummer des Sensors
		fmt.Println(u, sensornum)
		_, werte := sensordata.DecodeParsePackage(values, int(sensornum))
		fmt.Println(werte)
		sensordata.AddWertMult(int(sensornum), werte)

		lastTimes[int(sensornum)] = time.Now()

		return
	}

	w.WriteHeader(405)
	_, err := fmt.Fprintf(w, "NOT ALLOWED")

	if err != nil {
		Error.Println("error:", err)
		return
	}

}

func getHandler(w http.ResponseWriter, r *http.Request) {
	defer panicLogger()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Println(r.Host)
	fmt.Println(r.Method)
	fmt.Println(r.RequestURI)
	fmt.Println("GET")

	uri, err := url.ParseRequestURI(r.RequestURI)
	query := uri.Query()

	num, ok1 := query["n"]
	tim, ok2 := query["t"]
	del, ok3 := query["d"]

	ti, _ := strconv.Atoi(tim[0])
	delta, _ := strconv.Atoi(del[0])

	te := time.Now().Unix()
	ts := te - int64(ti)

	data := make([]sensordata.SensorWert, 0)

	if ok1 && ok2 && ok3 {

		for i := range num {
			senNumber, _ := strconv.Atoi(num[i])
			data = append(data, sensordata.Summit(senNumber, time.Unix(ts, 0), time.Unix(te, 0), delta)...)
		}

		bytes, _ := json.Marshal(data)
		_, err = w.Write(bytes)

		if err != nil {
			Error.Println("error:", err)
		}
		return
	}

	if err != nil {
		Error.Println("error:", err)
		return
	}
}

func keyHandler(w http.ResponseWriter, r *http.Request) {
	defer panicLogger()

	if r.Method == "GET" {
		//CREATE new Rand IDENTIFIER
		secure.FilterOld()
		newID := secure.AddNewIdent()
		b := newID.Id[:]
		fmt.Println(b)
		//Send IDENTIFIER
		_, err := w.Write(b)
		if err != nil {
			Error.Println("error:", err)
			return
		}
		return
	}

	w.WriteHeader(405)
	_, err := fmt.Fprintf(w, "NOT ALLOWED")
	if err != nil {
		Error.Println("error:", err)
		return
	}
}

func wakeHandler(w http.ResponseWriter, r *http.Request) {
	defer panicLogger()

	if r.Method == "PUT" {
		b := make([]byte, r.ContentLength)
		n, err := r.Body.Read(b)
		if n > 0 && err == nil {
			Warning.Printf("<h2>Sensor %v has Restarted</h2>", b[0])
		}
	}
}

func readMailAccountGob(filePath string, account *writemail.MailAccount) error {
	file, err := os.Open(filePath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(account)
	}
	err = file.Close()
	return err
}

func InitLogging() {

	account := writemail.MailAccount{}
	err := readMailAccountGob("account.gob", &account)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	msger := writemail.MailMsg{Account: account,
		Name:    "Creapolis Server Log",
		To:      []string{"oli1111@web.de", "oliver.walter@stud.hs-coburg.de", "Daniel.Melzer@stud.hs-coburg.de"},
		Subject: "Server Log"}

	Trace = log.New(msger, "<h1>Trace</h1>", log.LstdFlags|log.Llongfile)
	Info = log.New(msger, "<h1>Info</h1>", log.LstdFlags|log.Llongfile)
	Warning = log.New(msger, "<h1>Warning</h1>", log.LstdFlags|log.Llongfile)
	Error = log.New(msger, "<h1>Error</h1>", log.LstdFlags|log.Llongfile)

}

func CheckSensors() {

	for i := range lastTimes {
		if time.Since(lastTimes[i]) > 4*time.Hour {

			Error.Printf("<h2> Sensor %v Missing </h2> Last time online : %v \n Duration : %v ", i, lastTimes[i], time.Since(lastTimes[i]))

		}
	}

}

func main() {
	defer panicLogger()
	lastTimes = make([]time.Time, 5)
	for i := range lastTimes {
		lastTimes[i] = time.Now()
	}

	ticker := time.NewTicker(2 * time.Hour)

	go func() {
		for t := range ticker.C {
			fmt.Println("Tick at", t)
			CheckSensors()
		}
	}()

	sensordata.Init()
	InitLogging()
	Info.Printf("<h2>Sensor Log Server has Started</h2> Loaded %v datapoints from storage", sensordata.CountAll())

	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/put", putHandler)
	http.HandleFunc("/key", keyHandler)
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/wake", wakeHandler)
	err := http.ListenAndServe(":8000", nil)

	if err != nil {
		Error.Println("error:", err)
		return
	}
}
