package control

import (
	"net"
	"net/netip"
	"sync"

	E "github.com/sagernet/sing/common/exceptions"
)

var _ InterfaceFinder = (*DefaultInterfaceFinder)(nil)

type DefaultInterfaceFinder struct {
	mu sync.Mutex

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
	f.mu.Lock()
	f.interfaces = interfaces
	f.mu.Unlock()
	return nil
}

func (f *DefaultInterfaceFinder) UpdateInterfaces(interfaces []Interface) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.interfaces = interfaces
}

func (f *DefaultInterfaceFinder) Interfaces() []Interface {
	return f.interfaces
}

func (f *DefaultInterfaceFinder) ByName(name string) (*Interface, error) {
	for _, netInterface := range f.interfaces {
		if netInterface.Name == name {
			return &netInterface, nil
		}
	}
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
	for _, netInterface := range f.interfaces {
		if netInterface.Index == index {
			return &netInterface, nil
		}
	}
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
	for _, netInterface := range f.interfaces {
		for _, prefix := range netInterface.Addresses {
			if prefix.Contains(addr) {
				return &netInterface, nil
			}
		}
	}
	return nil, &net.OpError{Op: "route", Net: "ip+net", Source: nil, Addr: &net.IPAddr{IP: addr.AsSlice()}, Err: E.New("no such network interface")}
}
