package sensordata

import "fmt"

type SensorWert struct {
	Temperature float32
	Energy      float32
	Time        uint64
	Num         int
}

func (s SensorWert) Less(wert SensorWert) bool {
	return s.Time < wert.Time
}

func (s SensorWert) Add(wert SensorWert) SensorWert {
	s.Energy = s.Energy + wert.Energy
	s.Temperature = s.Temperature + wert.Temperature
	s.Time = s.Time + wert.Time
	if s.Num == -1 {
		s.Num = wert.Num
	} else {
		if s.Num != wert.Num {
			fmt.Println("ERROR Sensor number not Equal")
		}
	}
	return s
}

func (s SensorWert) DivScalar(div int) SensorWert {
	s.Temperature = s.Temperature / float32(div)
	//s.Energy = s.Energy / float32(div)
	s.Time = s.Time / uint64(div)
	return s
}

func MeanSlice(werte []SensorWert) SensorWert {
	sum := SensorWert{0, 0, 0, -1}

	for i := range werte {
		sum = sum.Add(werte[i])
	}
	sum.DivScalar(len(werte))
	return sum
}
