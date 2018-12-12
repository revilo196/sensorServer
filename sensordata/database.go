package sensordata

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

type Store [][]SensorWert



const filename = "database.gob"

var storage Store
var m sync.Mutex

func Init() {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
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
		m.Lock()
		err := readGob(filename)
		m.Unlock()
		if err != nil {
			fmt.Println(err)
			fmt.Println("Database Init Load Error")
		}
	}
}

func AddWertMult(num int, wert []SensorWert) {

	for i := range wert {
		AddWert(num, wert[i])
	}
}

func AddWert(num int, wert SensorWert) {

	if num > 5 || num < 0 {
		return
	}
	m.Lock()
	storage[num] = append(storage[num], wert)

	sort.Slice(storage[num], func(i, j int) bool {
		return storage[num][i].Time < storage[num][j].Time
	})
	m.Unlock()
	Save()
}

func EncodeSensor(num int) []byte {
	bytes, err := json.Marshal(storage[num])
	if err != nil {
		fmt.Println(err)
		fmt.Println("Convert Error")
	}
	return bytes
}

func GetPart(sensor int, timestart time.Time, timeend time.Time) []SensorWert {
	time_s := uint64(timestart.Unix()) + 3600
	time_e := uint64(timeend.Unix()) + 3600
	sen := GetSensor(sensor)
	//find index Start
	indexStart := sort.Search(len(sen)-1, func(i int) bool {
		if time_s < sen[i].Time {
			return true
		}
		return false
	})

	indexEnd := sort.Search(len(sen)-1, func(i int) bool {
		if time_e < sen[i].Time {
			return true
		}
		return false
	})

	if indexEnd < indexStart {
		s := indexEnd
		indexEnd = indexStart
		indexStart = s
	}

	return sen[indexStart:indexEnd]

}

func Summit(sensor int, timestart time.Time, timeend time.Time, deltaseconds int) []SensorWert {
	data := GetPart(sensor, timestart, timeend)

	var out []SensorWert
	sum := data[0]

	last := data[0].Time
	add := 1
	for i := 1; i < len(data); i++ {
		if data[i].Time-last < uint64(deltaseconds) {
			sum = sum.Add(data[i])
			add++
		} else {
			out = append(out, sum.DivScalar(add))
			add = 1
			last = data[i].Time
			sum = data[i]
		}
	}

	out = append(out, sum.DivScalar(add))
	return out
}

func GetSensor(num int) []SensorWert {
	return storage[num]
}

func Save() {
	err := writeGob(filename, storage)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Database Init Save Error")
	}
}

func CountAll() int {
	sum := 0
	for i := range storage {
		sum += len(storage[i])
	}
	return sum
}

func writeGob(filePath string, object interface{}) error {
	file, err := os.Create(filePath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		err = encoder.Encode(object)
	}
	file.Close()
	return err
}

func readGob(filePath string) error {
	file, err := os.Open(filePath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(&storage)
	}
	file.Close()
	return err
}
