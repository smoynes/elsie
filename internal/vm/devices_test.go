package vm

import (
	"testing"
)

var (
	_ Driver       = &Keyboard{} // A keyboard drives itself.
	_ DeviceReader = &Keyboard{}
	_ DeviceWriter = &Keyboard{}
)

func TestKeyboardDriver(tt *testing.T) {
	t := testHarness{tt}
	t.init()

	var (
		have = Device{
			status: 0x8000,
			data:   DeviceRegister('X'),
			driver: &Keyboard{},
		}
		want = Device{
			data:   0x0058,
			status: 0x0000,
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
	} else if DeviceRegister(got) != have.status {
		t.Errorf("read: status have: %s, got: %s", want.status, got)
	}

	got = Register(0xdad0)

	if err := mmio.Load(KBDRAddr, &got); err != nil {
		t.Error(err)
	} else if DeviceRegister(got) != want.data {
		t.Errorf("read: data want: %s, got: %s", want.data, got)
	}

	if err := mmio.Load(KBSRAddr, &got); err != nil {
		t.Error(err)
	} else if DeviceRegister(got) != want.status {
		t.Errorf("read: status want: %s, got: %s", want.status, got)
	}

	defer func() {
		if perr := recover(); perr != nil {
			t.Errorf("status: store panicked: %s", perr)
		} else {
			t.Logf("status: store did not panic: %s", have)
		}

	}()

	if err := mmio.Store(KBSRAddr, 0xf001); err != nil {
		t.Error(err)
	}
}