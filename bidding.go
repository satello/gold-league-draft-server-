package main

import (
  "log"
)

type Nomination struct {
  // player nominated
  player *Player

  // id of the bidder that nominated player
  bidderId string

  // amount of initial bid
  amount int
}

type Bid struct {
  // amount of the bid
  amount int

  // timestamp of bid
  // TODO do I want this?

  // who made the bid
  bidderId string
}

func broadcastNewBidderNominee(bidder *Bidder, h *DraftHub) {
  response := Response{"NEW_NOMINEE", map[string]interface{}{"bidderId": bidder.BidderId}}
  response_json := responseToJson(response)
  broadcastMessage(h, response_json)
}

func broadcastNewPlayerNominee(player *Player, h *DraftHub) {
  log.Println(player)
  response := Response{"NEW_PLAYER_NOMINEE", map[string]interface{}{"name": player.Name, "bid": player.bid.amount, "bidderId": player.bid.bidderId}}
  response_json := responseToJson(response)
  broadcastMessage(h, response_json)
}

func broadcastNewPlayerBid(player *Player, h *DraftHub) {
  response := Response{"NEW_PLAYER_BID", map[string]interface{}{"name": player.Name, "bid": player.bid.amount, "bidderId": player.bid.bidderId}}
  response_json := responseToJson(response)
  broadcastMessage(h, response_json)
}

func updateCountdown(ticks int, h *DraftHub) {
  response := Response{"TICKER_UPDATE", map[string]interface{}{"ticks": ticks}}
  response_json := responseToJson(response)
  broadcastMessage(h, response_json)
  log.Println("Just broadcasted")
}

func nextNomination(h *DraftHub) {
  // start next nomination cycle if one isn't already running
  if (!h.nominationCycle.open) {
    var curIndex int
    curIndex = h.curBidderIndex

    var nextNominator *Bidder
    firstCycle := true

    loop:
    for {
      // end draft if we nobody eligable to bid
      if !firstCycle && h.curBidderIndex == curIndex {
        // FIXME end draft properly
        log.Println("why is it getting here?")
        return
      }
      firstCycle = false
      nextNominator = h.biddersSlice[h.curBidderIndex]
      // if this nominator is allowed to keep drafting select them
      log.Println(nextNominator.Name)
      log.Println(nextNominator.Draftable)
      if nextNominator.Draftable {
        log.Println("should be first iteration %d", h.curBidderIndex)
        break loop
      } else {
        // if current bidder not eligable go to the next one
        h.curBidderIndex = (h.curBidderIndex + 1) % len(h.biddersSlice)
      }
    }

    h.draftState.CurrentNominatorId = nextNominator.BidderId
    broadcastNewBidderNominee(nextNominator, h)
    h.draftState.nominating = true
    go h.nominationCycle.getNominee(h, nextNominator.BidderId)
  }
}

func previousNomination(h *DraftHub, player *Player) {
  // FIXME this function pretty gross, error prone and hard to maintain
  // no bidding cycle can currently be running
  // start with previous index
  var prevIndex int
  // take off 1 from bidder index
  h.curBidderIndex -= 1
  prevIndex = h.curBidderIndex

  bidder := h.biddersMap[player.bid.bidderId]
  if bidder == nil {
    return
  }

  // return money, spot and eligibility
  bidder.Cap += player.bid.amount
  bidder.Spots += 1
  bidder.Draftable = true
  // reset bidder state
  broadcastBidderState(bidder, h)

  // looking for previous nominator
  var prevNominator *Bidder
  firstCycle := true

  loop:
  for {
    // end draft if we nobody eligable to bid
    if !firstCycle && h.curBidderIndex == prevIndex {
      // FIXME end draft properly
      return
    }
    firstCycle = false
    prevNominator = h.biddersSlice[h.curBidderIndex]

    if prevNominator.Draftable {
      break loop
    } else {
      // if current bidder not eligable go to the next one
      h.curBidderIndex = (h.curBidderIndex - 1) % len(h.biddersSlice)
    }
  }

  // reset player bids
  player.bid.bidderId = prevNominator.BidderId
  player.bid.amount = 1
  player.Taken = false
  // put player back on list
  braodcastPlayers(h)

  // set state
  // FIXME yuck
  h.draftState.CurrentNominatorId = prevNominator.BidderId
  h.draftState.CurrentBidderId = prevNominator.BidderId
  h.draftState.CurrentBid = 1
  h.draftState.CurrentPlayerName = player.Name
  h.draftState.Paused = true
  braodcastDraftState(h)
  broadcastNewBidderNominee(prevNominator, h)

  // start bid cycle
  h.startBidding <- player
  log.Println("pause so we are g2g")
  h.biddingCycle.pauseChan <- true
}
