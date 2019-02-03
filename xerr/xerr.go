package xerr

import (
	"fmt"
	"net"
	"strings"
)

type Error struct {
	Err  error
	Kind Kind
}

func (e *Error) Error() string {
	var b strings.Builder
	if e.Kind != 0 {
		b.WriteString(e.Kind.String())
	}
	if b.Len() != 0 {
		b.WriteString(": ")
	}
	fmt.Fprintf(&b, "%v", e.Err)
	return b.String()
}

func (e *Error) Temporary() bool  { return e.Kind&Temporary != 0 }
func (e *Error) Timeout() bool    { return e.Kind&Timeout != 0 }
func (e *Error) Permission() bool { return e.Kind&Permission != 0 }

func Wrap(err error, k Kind) error {
	return &Error{Err: err, Kind: k}
}

func Is(err error, k Kind) bool {
	if e, ok := err.(*Error); ok {
		if e.Kind&k == k {
			return true
		}
		if e.Err != nil {
			return Is(e.Err, k)
		}
		return false
	}
	return k.matches(err)
}

const (
	Timeout    Kind = 1 << iota // see: timeout
	Temporary                   // see: temporary
	Permission                  // see: permission
)

type Kind uint

var _ error = Kind(0)

func (k Kind) Error() string { return k.String() }
func (k Kind) String() string {
	if k == 0 {
		return ""
	}
	var b strings.Builder
	for i := Kind(1); k != 0; i <<= 1 {
		if k&i == 0 {
			continue
		}
		switch k ^= i; i {
		case Timeout:
			b.WriteString("timeout, ")
		case Temporary:
			b.WriteString("temporary, ")
		case Permission:
			b.WriteString("permission, ")
		default:
			fmt.Fprintf(&b, "Kind(%d)", int(i))
		}
	}
	return b.String()[:b.Len()-2] // trim ", "
}

type timeout interface{ Timeout() bool }
type temporary interface{ Temporary() bool }
type permission interface{ Permission() bool }

func (k Kind) matches(v interface{}) bool {
	for i := Kind(1); k != 0; i <<= 1 {
		if k&i == 0 {
			continue
		}
		switch k ^= i; i {
		case Temporary:
			if t, ok := v.(temporary); !ok || !t.Temporary() {
				return false
			}
		case Timeout:
			if t, ok := v.(timeout); !ok || !t.Timeout() {
				return false
			}
		case Permission:
			if p, ok := v.(permission); !ok || !p.Permission() {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// For ../poll[1,N]

type ReauthError interface {
	error
	Permission() bool
	Temporary() bool
}

type TimeoutError interface {
	net.Error
}

// Is implements the unexported errors.Is interface.
func (e *Error) Is(err error) bool {
	if k, ok := err.(Kind); ok {
		return k&e.Kind == e.Kind
	}
	return false
}
