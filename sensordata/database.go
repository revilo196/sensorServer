package sensordata

import (
	"encoding/gob"
	"fmt"
	"os"
	"sort"
)

type Store [][]SensorWert

const filename = "database.gob"

var storage Store

func Init() {
	_, err := os.Stat(filename)

	if !os.IsNotExist(err) {
		storage = make([][]SensorWert, 5)

		for i := range storage {
			storage[i] = make([]SensorWert, 1)
		}

		err := writeGob(filename, storage)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Database Init Save Error")
		}
	} else {
		err := readGob(filename, storage)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Database Init Load Error")
		}
	}
}

func AddWert(num int, wert SensorWert) {

	if num > 5 || num < 0 {
		return
	}
	storage[num] = append(storage[num], wert)

	sort.Slice(storage[num], func(i, j int) bool {
		return storage[num][i].Time < storage[num][j].Time
	})

}

func writeGob(filePath string, object interface{}) error {
	file, err := os.Create(filePath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

func readGob(filePath string, object interface{}) error {
	file, err := os.Open(filePath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}
