package bus

import (
	"log"
	"testing"

	"github.com/jsmorph/evpat/pat"
)

func TestRing(t *testing.T) {
	var (
		r  = NewRing(5)
		fi = 0
		f  = func() *Msg {
			fi++
			return &Msg{
				Payload: fi,
			}
		}
		show = func(msgs []*Msg) {
			log.Println()
			for i, msg := range msgs {
				log.Printf("%d %s", i, pat.JSON(msg))
			}
		}
	)

	for i := 0; i < r.size*3; i++ {
		r.add(f())
		show(r.ReplayRecent(3))
	}
}
