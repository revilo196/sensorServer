package main

import (
	"fmt"
	"golang.org/x/crypto/blake2s"
	"io"
	"net/http"
	"sensorServer/secure"
	"sensorServer/sensordata"
)

var key = []byte{0x2b, 0x7e, 0x15, 0x16, 0x28, 0xae, 0xd2, 0xa6, 0xab, 0xf7, 0x15, 0x88, 0x09, 0xcf, 0x4f, 0x3c,
	0x3e, 0x05, 0xb6, 0x96, 0x55, 0xea, 0x2e, 0xae, 0xe9, 0xee, 0xf1, 0xa2, 0x2f, 0x13, 0x39, 0x99}

func putHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "PUT" {

		//#READ INPUT
		b := make([]byte, r.ContentLength)
		n, err := r.Body.Read(b)

		if err != nil && err != io.EOF {
			fmt.Println("error:", err, n)
			fmt.Println(b)
			return
		}

		//#DECRYPT
		plaintext := secure.Decrypt(b)

		//#DECODE
		u := plaintext[:15]
		sensornum := plaintext[15]
		prufsumme := plaintext[len(plaintext)-16:]
		pack := plaintext[:len(plaintext)-16]
		values := pack[16:]

		hash, err := blake2s.New128(key[:16])
		hash.Write(pack)
		bs := hash.Sum(nil)

		for i := range bs {
			if bs[i] != prufsumme[i] {
				fmt.Fprintf(w, "HASH")
				return
			}
		}

		// U: IDENTIFIER  |  sensornum: Nummer des Sensors  |  f1,f2 Sensor Werte
		fmt.Println(u, sensornum)

		valid := secure.CheckID(u)

		fmt.Println(valid)
		if valid {
			fmt.Fprintf(w, "OK")
			_, werts := sensordata.DecodeParsePackage(values)
			fmt.Println(werts)
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
