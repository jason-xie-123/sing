package control

import (
	"net"
	"net/netip"
	"sync"

	E "github.com/sagernet/sing/common/exceptions"
)

var _ InterfaceFinder = (*DefaultInterfaceFinder)(nil)

type DefaultInterfaceFinder struct {
	interfacesAccess sync.Mutex

	interfaces []Interface
}

func NewDefaultInterfaceFinder() *DefaultInterfaceFinder {
	return &DefaultInterfaceFinder{}
}

func (f *DefaultInterfaceFinder) Update() error {
	netIfs, err := net.Interfaces()
	if err != nil {
		return err
	}
	interfaces := make([]Interface, 0, len(netIfs))
	for _, netIf := range netIfs {
		var iif Interface
		iif, err = InterfaceFromNet(netIf)
		if err != nil {
			return err
		}
		interfaces = append(interfaces, iif)
	}
	f.interfacesAccess.Lock()
	f.interfaces = interfaces
	f.interfacesAccess.Unlock()
	return nil
}

func (f *DefaultInterfaceFinder) UpdateInterfaces(interfaces []Interface) {
	f.interfacesAccess.Lock()
	defer f.interfacesAccess.Unlock()
	f.interfaces = interfaces
}

func (f *DefaultInterfaceFinder) Interfaces() []Interface {
	f.interfacesAccess.Lock()
	defer f.interfacesAccess.Unlock()
	var interfaces []Interface
	interfaces = append(interfaces, f.interfaces...)
	return interfaces
}

func (f *DefaultInterfaceFinder) ByName(name string) (*Interface, error) {
	f.interfacesAccess.Lock()
	for _, netInterface := range f.interfaces {
		if netInterface.Name == name {
			defer f.interfacesAccess.Unlock()
			return &netInterface, nil
		}
	}
	defer f.interfacesAccess.Unlock()

	_, err := net.InterfaceByName(name)
	if err == nil {
		err = f.Update()
		if err != nil {
			return nil, err
		}
		return f.ByName(name)
	}
	return nil, &net.OpError{Op: "route", Net: "ip+net", Source: nil, Addr: &net.IPAddr{IP: nil}, Err: E.New("no such network interface")}
}

func (f *DefaultInterfaceFinder) ByIndex(index int) (*Interface, error) {
	f.interfacesAccess.Lock()
	for _, netInterface := range f.interfaces {
		if netInterface.Index == index {
			f.interfacesAccess.Unlock()
			return &netInterface, nil
		}
	}
	f.interfacesAccess.Unlock()

	_, err := net.InterfaceByIndex(index)
	if err == nil {
		err = f.Update()
		if err != nil {
			return nil, err
		}
		return f.ByIndex(index)
	}
	return nil, &net.OpError{Op: "route", Net: "ip+net", Source: nil, Addr: &net.IPAddr{IP: nil}, Err: E.New("no such network interface")}
}

func (f *DefaultInterfaceFinder) ByAddr(addr netip.Addr) (*Interface, error) {
	f.interfacesAccess.Lock()
	for _, netInterface := range f.interfaces {
		for _, prefix := range netInterface.Addresses {
			if prefix.Contains(addr) {
				f.interfacesAccess.Unlock()
				return &netInterface, nil
			}
		}
	}
	f.interfacesAccess.Unlock()

	return nil, &net.OpError{Op: "route", Net: "ip+net", Source: nil, Addr: &net.IPAddr{IP: addr.AsSlice()}, Err: E.New("no such network interface")}
}
