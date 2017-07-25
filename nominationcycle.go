package main

import (
  "time"
)

type NominationCycle struct {
  // message channel for new nominations
  nominationChan chan *Nomination

  open bool
}

func newNominationCycle() *NominationCycle {

	return &NominationCycle{
    nominationChan: make(chan *Nomination),
    open: false,
	}
}

// use as go routine. has callback to hub
func (d *NominationCycle) getNominee(h *DraftHub) {
  d.open = true
  ticks := 30
  nominationTicker := time.NewTicker(time.Second)

  loop:
  for {
    select {
    case <- nominationTicker.C:
      ticks -= 1
      updateCountdown(ticks, h)
      if ticks < 1 {
        nominationTicker.Stop()
        // TODO handle person not nominating someone in time
        h.startBidding <- &Player{
          Name: "shit stain",
        }
        d.open = false
        break loop
      }
    case nomination := <- d.nominationChan:
      nominationTicker.Stop()
      currentPlayer := nomination.player

      currentPlayer.bid = &Bid{
        amount: 1,
        bidderId: nomination.bidderId,
      }
      // call back to hub that you have a new player up for bid
      h.startBidding <- currentPlayer
      d.open = false
      break loop
    }
  }
}
