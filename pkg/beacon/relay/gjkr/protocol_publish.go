package gjkr

import (
	"fmt"
	"math"
	"math/big"

	"github.com/keep-network/keep-core/pkg/beacon/relay/result"
)

// PrepareResult sets results of distributed key generation. It takes generated
// group public key along with disqualified and inactive members and stores
// in member's result field.
//
// Additional validation to check if number of disqualified and inactive members
// is greater than half of the configured dishonest threshold. If so the group
// is to weak and the result is set to a failure.
func (pm *PublishingMember) PrepareResult() {
	group := pm.group
	disqualifiedMembers := group.DisqualifiedMembers()
	inactiveMembers := group.InactiveMembers()

	// if nPlayers(IA + DQ) > T/2:
	if len(disqualifiedMembers)+len(inactiveMembers) > (group.dishonestThreshold / 2) {
		// Result.failure(disqualified = DQ)
		pm.result = &result.Result{
			Success:      false,
			Disqualified: disqualifiedMembers,
		}
	} else {
		// Result.success(pubkey = Y, inactive = IA, disqualified = DQ)
		pm.result = &result.Result{
			Success:        true,
			GroupPublicKey: big.NewInt(123), // TODO: Use group public key after Phase 12 is merged
			Disqualified:   disqualifiedMembers,
			Inactive:       inactiveMembers,
		}
	}
}

// determinePublishersIDs determines IDs of members eligable to submit the result
// to the blockchain. It takes into a consideration the number of blocks which has
// passed during the protocol execution. If protocol execution time did not
// exceed expected protocol duration, then first group member is eligable to
// publish the result. If expected protocol duration is exceeded, then next members
// are added to the eligable publishers. Subsequent members are added to the group
// as the blocks defined by `blockStep` pass.
func (pm *PublishingMember) determinePublishersIDs() ([]int, error) {
	expectedProtocolDuration := pm.protocolConfig.chain.expectedProtocolDuration // t_dkg

	// Current block height.
	currentBlock, err := pm.protocolConfig.chain.CurrentBlock() // t_now
	if err != nil {
		return nil, fmt.Errorf("getting current block height failed [%v]", err)
	}

	// Time elapsed from protocol execution initialization.
	// `T_elapsed = T_now - T_init`
	elapsedBlocks := currentBlock - pm.protocolConfig.chain.initialBlockHeight

	// Determine highest member index eligible to publish the result.
	var highestMemberIndex int // j
	// If elapsed time is less than expected protocol execution duration.
	if elapsedBlocks <= expectedProtocolDuration { // if T_elapsed <= T_dkg
		highestMemberIndex = 0 // in protocol spec first player is denoted as `j=1`
	} else {
		// Current execution time exceeded expected protocol execution duration.
		surpassBlocks := elapsedBlocks - expectedProtocolDuration // T_over = T_elapsed - T_dkg
		// j = 1 + ceiling(T_over / T_step)
		highestMemberIndex = int(math.Ceil(float64(surpassBlocks / pm.protocolConfig.chain.blockStep)))
	}

	// Select group members with index less or equal the highest member index.
	var publishersIDs []int
	for index, groupMemberID := range pm.group.MemberIDs() {
		if index <= highestMemberIndex { // if j >= i
			publishersIDs = append(publishersIDs, groupMemberID)
		} else {
			break
		}
	}

	return publishersIDs, nil
}
