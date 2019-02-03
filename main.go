// +build ignore

package main

import (
	"time"

	"github.com/ericlagergren/error-example/errors"
	"github.com/ericlagergren/error-example/xerr"
)

func main() {}

func someAPIQuery() error { return nil }
func someAPIReauth()      {}

func poll0() error {
	for i := 0; ; i++ {
		switch err := someAPIQuery(); {
		case xerr.Is(err, xerr.Permission|xerr.Temporary):
			someAPIReauth()
		case xerr.Is(err, xerr.Timeout|xerr.Temporary):
			<-time.After(time.Duration(i) * time.Second)
		default:
			return err
		}
	}
}

func poll1() error {
	for i := 0; ; i++ {
		switch err := someAPIQuery(); {
		case errors.Is(err, xerr.Permission|xerr.Temporary):
			someAPIReauth()
		case errors.Is(err, xerr.Timeout|xerr.Temporary):
			<-time.After(time.Duration(i) * time.Second)
		default:
			return err
		}
	}
}

func poll2() error {
	for i := 0; ; i++ {
		err := someAPIQuery()

		var reauth xerr.ReauthError
		if errors.Is(err, reauth) &&
			err.(xerr.ReauthError).Permission() &&
			err.(xerr.ReauthError).Temporary() {
			someAPIReauth()
			continue
		}

		var retry xerr.TimeoutError
		if errors.Is(err, retry) &&
			err.(xerr.TimeoutError).Timeout() &&
			err.(xerr.TimeoutError).Temporary() {
			<-time.After(time.Duration(i) * time.Second)
			continue
		}

		return err
	}
}

func poll3() error {
	for i := 0; ; i++ {
		err := someAPIQuery()

		var reauth xerr.ReauthError
		if errors.As(err, &reauth) && reauth.Permission() && reauth.Temporary() {
			someAPIReauth()
			continue
		}

		var retry xerr.TimeoutError
		if errors.As(err, &retry) && retry.Timeout() && retry.Temporary() {
			<-time.After(time.Duration(i) * time.Second)
			continue
		}

		return err
	}
}
