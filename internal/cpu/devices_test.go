package cpu

import (
	"testing"
)

var (
	_ Driver       = KeyboardDriver{}
	_ DeviceReader = KeyboardDriver{}
	_ DeviceWriter = KeyboardDriver{}
)

func TestKeyboardDriver(tt *testing.T) {
	t := testHarness{tt}
	t.init()

	var (
		have = Device{
			status: 0x0000,
			data:   DeviceRegister('X'),
			driver: KeyboardDriver{},
		}
		want = Device{
			data: 0x0058,
		}
	)

	mmio := MMIO{}

	err := mmio.Map(map[Word]any{
		KBSRAddr: &have,
		KBDRAddr: &have,
	})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	got := Register(0xface)

	if err := mmio.Load(KBSRAddr, &got); err != nil {
		t.Error(err)
	} else if KeyboardStatus(got) != KeyboardStatus(have.status) {
		t.Errorf("Status have: %s, got: %s", want.status, got)
	}

	got = Register(0xdad0)

	if err := mmio.Load(KBDRAddr, &got); err != nil {
		t.Error(err)
	} else if KeyboardData(got) != KeyboardData(want.data) {
		t.Errorf("Data want: %s, got: %s", want.data, got)
	}

	if err := mmio.Load(KBSRAddr, &got); err != nil {
		t.Error(err)
	} else if KeyboardData(got) != KeyboardData(want.status) {
		t.Errorf("Data want: %s, got: %s", want.data, got)
	}

	defer func() {
		if perr := recover(); perr != nil {
			t.Errorf("Store panicked: %s", perr)
		} else {
			t.Logf("Store did not panic: %s", have)
		}

	}()

	if err := mmio.Store(KBSRAddr, 0xf001); err != nil {
		t.Error(err)
	}
}
