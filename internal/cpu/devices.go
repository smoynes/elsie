package cpu

// devices.go has device drivers.

type Keyboard struct {
	status KeyboardStatus
	data   KeyboardData
}
type KeyboardStatus Register

func (k KeyboardStatus) String() string {
	return Register(k).String()
}

type KeyboardData Register

func (k KeyboardData) String() string {
	return Register(k).String()
}

type Display struct {
	status DisplayStatus
	data   DisplayData
}
type DisplayStatus Register

func (k DisplayStatus) String() string {
	return Register(k).String()
}

type DisplayData Register

func (k DisplayData) String() string {
	return Register(k).String()
}
