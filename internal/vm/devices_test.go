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
	t := NewTestHarness(tt)
	vm := t.Make()

	var (
		kbd = &Keyboard{
			Device: Device{
				data:   '?',
				status: 0x8000,
			},
		}
	)

	var (
		have = newDevice(vm, kbd, nil)
		want = Device{
			data:   0x003f,
			status: 0x0000,
		}
	)

	have.driver.(*Keyboard).WithLogger(t.log)

	mmio := NewMMIO()
	mmio.WithLogger(t.log)

	err := mmio.Map(map[Word]any{
		KBSRAddr: &have,
		KBDRAddr: &have,
	})

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	got := Register(0xface)

	have.driver.(*Keyboard).WithLogger(t.log)
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
		t.log.Panicf("load: data: kbd: %s", kbd)
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
