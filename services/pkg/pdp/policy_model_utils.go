package pdp

import common_models "github.com/abhisek/supply-chain-gateway/services/pkg/common/models"

func NewPolicyInputWithArtefact(target common_models.Artefact) PolicyInput {
	return PolicyInput{
		Kind: policyInputKind,
		Version: PolicyInputVersion{
			Major: policyInputMajorVersion,
			Minor: policyInputMinorVersion,
			Patch: policyInputPatchVersion,
		},
		Target: PolicyInputTarget{
			Artefact: PolicyEvalTargetArtefact{target},
		},
	}
}

func (s PolicyResponse) Allowed() bool {
	return (s.Allow) && (len(s.Violations) == 0)
}