package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"time"
)

var key = []byte{0x2b, 0x7e, 0x15, 0x16, 0x28, 0xae, 0xd2, 0xa6, 0xab, 0xf7, 0x15, 0x88, 0x09, 0xcf, 0x4f, 0x3c,
	0x3e, 0x05, 0xb6, 0x96, 0x55, 0xea, 0x2e, 0xae, 0xe9, 0xee, 0xf1, 0xa2, 0x2f, 0x13, 0x39, 0x99}

type identi struct {
	id   [23]byte
	time time.Time
}

func (ident identi) equalID(id []byte) bool {
	if len(id) != 23 {
		return false
	}
	for i := range id {
		if id[i] != ident.id[i] {
			return false
		}
	}
	return true
}

const TIMEOUT float64 = 30.0

func (ident identi) timeValid() bool {
	return time.Now().Sub(ident.time).Seconds() < TIMEOUT
}

type identList []identi

var ids = identList{}

func (ids *identList) filterTimeouts() {

}

func (ids identList) containsValid(id []byte) (bool, int) {
	for i := range ids {
		if ids[i].equalID(id) && ids[i].timeValid() {

			return true, i
		}
	}
	return false, -1
}

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
		ciphertext := b
		block, err := aes.NewCipher(key)
		if err != nil {
			panic(err)
		}
		iv := ciphertext[:aes.BlockSize]
		ciphertext = ciphertext[aes.BlockSize:]
		mode := cipher.NewCBCDecrypter(block, iv)
		mode.CryptBlocks(ciphertext, ciphertext)
		fmt.Println(ciphertext)

		//#DECODE
		u := ciphertext[:23]
		sensornum := ciphertext[23]
		buf1 := bytes.NewReader(ciphertext[24:28])
		buf2 := bytes.NewReader(ciphertext[28:32])
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

		bo, idx := ids.containsValid(u)
		if bo {
			ids = ids[:idx+copy(ids[idx:], ids[idx+1:])]
		}
		fmt.Println(bo)

		fmt.Fprintf(w, "OK")
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
		c := aes.BlockSize*2 - 9
		b := make([]byte, c)
		_, err := rand.Read(b)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		//Send IDENTIFIER
		fmt.Println(bytes.Equal(b, make([]byte, c)))
		_, err = w.Write(b)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(b)

		//ADD IDENTIFIER to Valid ID-List
		newId := identi{}
		copy(newId.id[:], b)
		newId.time = time.Now()
		ids = append(ids, newId)
		fmt.Println(ids)
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
