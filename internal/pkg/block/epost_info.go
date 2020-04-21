package block

import (
	"github.com/filecoin-project/specs-actors/actors/abi"
)

// EPoStCandidate wraps the input data needed to verify an election PoSt
type EPoStCandidate struct {
	_                    struct{} `cbor:",toarray"`
	PartialTicket        abi.PartialTicket
	SectorID             abi.SectorNumber
	SectorChallengeIndex int64
}

type EPoStProof struct {
	_               struct{} `cbor:",toarray"`
	RegisteredProof abi.RegisteredProof
	ProofBytes      []byte
}

// NewEPoStCandidate constructs an epost candidate from data
func NewEPoStCandidate(sID uint64, pt []byte, sci int64) EPoStCandidate {
	return EPoStCandidate{
		SectorID:             abi.SectorNumber(sID),
		PartialTicket:        pt,
		SectorChallengeIndex: sci,
	}
}

// NewEPoStProof constructs an epost proof from registered proof and bytes
func NewEPoStProof(rpp abi.RegisteredProof, bs []byte) EPoStProof {
	return EPoStProof{
		RegisteredProof: rpp,
		ProofBytes:      bs,
	}
}

// FromFFICandidate converts a Candidate to an EPoStCandidate
func FromFFICandidate(candidate abi.PoStCandidate) EPoStCandidate {
	return EPoStCandidate{
		PartialTicket:        candidate.PartialTicket[:],
		SectorID:             candidate.SectorID.Number,
		SectorChallengeIndex: candidate.ChallengeIndex,
	}
}

// FromFFICandidates converts a variable number of Candidate to a slice of
// EPoStCandidate
func FromFFICandidates(candidates ...abi.PoStCandidate) []EPoStCandidate {
	out := make([]EPoStCandidate, len(candidates))
	for idx, c := range candidates {
		out[idx] = FromFFICandidate(c)
	}

	return out
}

// FromABIPoStProofs converts the abi post proof type to a local type for
// serialization purposes
func FromABIPoStProofs(postProofs ...abi.PoStProof) []EPoStProof {
	out := make([]EPoStProof, len(postProofs))
	for i, p := range postProofs {
		out[i] = EPoStProof{RegisteredProof: p.RegisteredProof, ProofBytes: p.ProofBytes}
	}

	return out
}
