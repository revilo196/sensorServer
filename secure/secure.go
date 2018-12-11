package secure

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/blake2s"
	"time"
)

var key = []byte{0x2b, 0x7e, 0x15, 0x16, 0x28, 0xae, 0xd2, 0xa6, 0xab, 0xf7, 0x15, 0x88, 0x09, 0xcf, 0x4f, 0x3c,
	0x3e, 0x05, 0xb6, 0x96, 0x55, 0xea, 0x2e, 0xae, 0xe9, 0xee, 0xf1, 0xa2, 0x2f, 0x13, 0x39, 0x99}

type Ident struct {
	Id   [15]byte
	time time.Time
}

func (ident Ident) equalID(id []byte) bool {
	if len(id) != 15 {
		return false
	}
	for i := range id {
		if id[i] != ident.Id[i] {
			return false
		}
	}
	return true
}

const TIMEOUT float64 = 30.0

func (ident Ident) timeValid() bool {
	return time.Now().Sub(ident.time).Seconds() < TIMEOUT
}

type identList []Ident

var ids = identList{}

func (ids *identList) filterTimeouts() {
	for i := range *ids {
		if !(*ids)[i].timeValid() {
			*ids = (*ids)[:i+copy((*ids)[i:], (*ids)[i+1:])]
			ids.filterTimeouts()
			return
		}
	}
}

func (ids *identList) containsValid(id []byte) (bool, int) {
	for i := range *ids {
		if (*ids)[i].equalID(id) && (*ids)[i].timeValid() {
			*ids = (*ids)[:i+copy((*ids)[i:], (*ids)[+1:])]
			return true, i
		}
	}
	return false, -1
}

func FilterOld() {
	ids.filterTimeouts()
	return
}

func CheckID(id []byte) bool {
	ids.filterTimeouts()
	bo, _ := ids.containsValid(id)
	return bo
}

func AddNewIdent() Ident {
	c := aes.BlockSize - 1
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
	}

	newId := Ident{}
	copy(newId.Id[:], b)
	newId.time = time.Now()
	ids = append(ids, newId)
	fmt.Println(len(ids))
	return newId
}

func Decrypt(ciphertext []byte) []byte {

	//#DECRYPT
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCTR(block, iv)

	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext

}

func CheckHash(pack []byte) bool {
	sumA := pack[len(pack)-16:]
	pack = pack[:len(pack)-16]
	hash, err := blake2s.New128(key[:16])
	hash.Write(pack)
	sumB := hash.Sum(nil)

	if err != nil {
		return false
	}

	return bytes.Equal(sumA, sumB)

}
