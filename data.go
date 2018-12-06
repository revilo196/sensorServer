package main

type Svalue struct {
	 date uint32
	 temp float32
	 power float32
}

func (s Svalue) Less(b Svalue) bool {
	return s.date < b.date
}

