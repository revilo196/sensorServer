package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"sensorServer/secure"
)

func putHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "PUT" && r.ContentLength == 48 {

		//#READ INPUT
		b := make([]byte, 48)
		n, err := r.Body.Read(b)

		if err != nil && err != io.EOF {
			fmt.Println("error:", err, n)
			fmt.Println(b)
			return
		}
		fmt.Println(b)

		//#DECRYPT
		plaintext := secure.Decrypt(b)
		fmt.Println(plaintext)

		//#DECODE
		u := plaintext[:23]
		sensornum := plaintext[23]
		buf1 := bytes.NewReader(plaintext[24:28])
		buf2 := bytes.NewReader(plaintext[28:32])
		var f1, f2 float32

		err = binary.Read(buf1, binary.LittleEndian, &f1)
		if err != nil {
			fmt.Println("binary.Read failed:", err)
			fmt.Fprintf(w, "EV1")
			return
		}
		err = binary.Read(buf2, binary.LittleEndian, &f2)
		if err != nil {
			fmt.Println("binary.Read failed:", err)
			fmt.Fprintf(w, "EV2")
			return
		}

		// U: IDENTIFIER  |  sensornum: Nummer des Sensors  |  f1,f2 Sensor Werte
		fmt.Println(u, sensornum, f1, f2)

		valid := secure.CheckID(u)

		fmt.Println(valid)
		if valid {
			fmt.Fprintf(w, "OK")
		} else {
			fmt.Fprintf(w, "DENIED")

		}
		return
	}
	w.WriteHeader(405)
	_, err := fmt.Fprintf(w, "NOT ALLOWED")

	if err != nil {
		fmt.Println("error:", err)
		return
	}

}

func getHandler(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "<h1>Hello Internet</h1>")
	fmt.Println(r.Host)
	fmt.Println(r.Method)
	fmt.Println(r.ContentLength)

	fmt.Println("GET")

	if err != nil {
		fmt.Println("error:", err)
		return
	}

}

func keyHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		//CREATE new Rand IDENTIFIER
		secure.FilterOld()
		newID := secure.AddNewIdent()
		b := newID.Id[:]
		fmt.Println(b)
		//Send IDENTIFIER
		_, err := w.Write(b)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		return
	}

	w.WriteHeader(405)
	_, err := fmt.Fprintf(w, "NOT ALLOWED")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
}

func main() {

	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/put", putHandler)
	http.HandleFunc("/key", keyHandler)
	http.HandleFunc("/get", getHandler)
	err := http.ListenAndServe(":8000", nil)

	if err != nil {
		fmt.Println("error:", err)
		return
	}
}
