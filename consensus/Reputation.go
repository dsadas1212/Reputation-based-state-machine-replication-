// this file finish how to compute node's reputation
package consensus

import (
	"math"
	"sync"

	"github.com/adithyabhatkajake/libchatter/log"
)

const (
	pEpsilonWith  = 1.5
	pEpsilonEqui  = 1.5
	pEpsilonMali  = 0.5
	vEpisilonMali = 0.5
	gama          = 0.5
)

var (
	proposalnum     uint64 = 0
	votenum         uint64 = 0
	maliproposalnum uint64 = 0
	equiprospoalnum uint64 = 0
	withpropsoalnum uint64 = 0
	malivotenum     uint64 = 0
)

func (n *SyncHS) ReputationCalculateinCurrentRound(nodeID uint64) {
	//first we get the correct proposal/vote from map
	//get current various proposal/vote number
	// roundnum := n.view
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		n.proposalNumCalculate(nodeID)
		n.voteNumCalculate(nodeID)
		n.maliproposalNumCalculate(nodeID)
		n.equivocationproposalNumCalculate(nodeID)
		n.withholdproposalNumCalculate(nodeID)
		n.malivoteNumCalculate(nodeID)

	}()
	wg.Wait()
	log.Info("calculate reputation for node", nodeID)
	proposalsc := float64(proposalnum) - (float64(maliproposalnum)*pEpsilonMali +
		float64(equiprospoalnum)*pEpsilonEqui +
		float64(withpropsoalnum)*pEpsilonWith)
	proposalscore := n.maxvaluecheck(proposalsc)
	votesc := float64(votenum) - float64(malivotenum)*vEpisilonMali
	votescore := n.maxvaluecheck(votesc)
	nodeScore := math.Tanh(gama * (votescore + proposalscore))
	log.Info("The reputation of", nodeID, "is", nodeScore)

}

func (n *SyncHS) proposalNumCalculate(nodeID uint64) uint64 {
	n.propMapLock.RLock()
	defer n.propMapLock.RUnlock()
	for _, senderMap := range n.proposalMap[n.GetID()] {

		if senderMap[nodeID] == 1 {
			proposalnum++
		}

	}
	return proposalnum
}
func (n *SyncHS) voteNumCalculate(nodeID uint64) uint64 {
	n.voteMapLock.RLock()

	defer n.voteMapLock.RUnlock()
	for _, senderMap := range n.voteMap[n.GetID()] {

		if senderMap[nodeID] == 1 {
			votenum++
		}

	}
	return votenum
}

func (n *SyncHS) maliproposalNumCalculate(nodeID uint64) uint64 {
	n.malipropLock.RLock()

	defer n.malipropLock.RUnlock()
	for _, senderMap := range n.maliproposalMap[n.GetID()] {

		if senderMap[nodeID] == 1 {
			maliproposalnum++
		}

	}
	return maliproposalnum
}

func (n *SyncHS) withholdproposalNumCalculate(nodeID uint64) uint64 {
	n.withpropoLock.RLock()
	defer n.withpropoLock.RUnlock()
	for _, senderMap := range n.withproposalMap[n.GetID()] {

		if senderMap[nodeID] == 1 {
			withpropsoalnum++
		}

	}
	return withpropsoalnum

}

func (n *SyncHS) equivocationproposalNumCalculate(nodeID uint64) uint64 {
	n.equipropLock.RLock()
	defer n.equipropLock.RUnlock()
	for _, senderMap := range n.equiproposalMap[n.GetID()] {

		if senderMap[nodeID] == 1 {
			equiprospoalnum++
		}

	}
	return equiprospoalnum
}

func (n *SyncHS) malivoteNumCalculate(nodeID uint64) uint64 {
	n.voteMaliLock.RLock()
	defer n.voteMaliLock.RUnlock()
	for _, senderMap := range n.voteMaliMap[n.GetID()] {

		if senderMap[nodeID] == 1 {
			malivotenum++
		}

	}
	return malivotenum
}

func (n *SyncHS) maxvaluecheck(a float64) float64 {
	if a >= 0 {
		return a
	} else {
		return 0
	}
}

//calcute the score of each node in this round
