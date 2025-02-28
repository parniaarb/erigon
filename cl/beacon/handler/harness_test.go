package handler_test

import (
	"testing"

	"github.com/ledgerwatch/erigon/cl/beacon/beacontest"
	"github.com/ledgerwatch/erigon/cl/clparams"

	_ "embed"
)

func TestHarnessPhase0(t *testing.T) {
	beacontest.Execute(
		append(
			defaultHarnessOpts(harnessConfig{t: t, v: clparams.Phase0Version}),
			beacontest.WithTestFromFs(Harnesses, "blocks"),
			beacontest.WithTestFromFs(Harnesses, "config"),
			beacontest.WithTestFromFs(Harnesses, "headers"),
			beacontest.WithTestFromFs(Harnesses, "attestation_rewards_phase0"),
			beacontest.WithTestFromFs(Harnesses, "committees"),
			beacontest.WithTestFromFs(Harnesses, "duties_attester"),
			beacontest.WithTestFromFs(Harnesses, "duties_proposer"),
		)...,
	)
}
func TestHarnessPhase0Finalized(t *testing.T) {
	beacontest.Execute(
		append(
			defaultHarnessOpts(harnessConfig{t: t, v: clparams.Phase0Version, finalized: true}),
			beacontest.WithTestFromFs(Harnesses, "liveness"),
			beacontest.WithTestFromFs(Harnesses, "duties_attester_f"),
			beacontest.WithTestFromFs(Harnesses, "committees_f"),
		)...,
	)
}

func TestHarnessBellatrix(t *testing.T) {
	beacontest.Execute(
		append(
			defaultHarnessOpts(harnessConfig{t: t, v: clparams.BellatrixVersion, finalized: true}),
			beacontest.WithTestFromFs(Harnesses, "attestation_rewards_bellatrix"),
			beacontest.WithTestFromFs(Harnesses, "duties_sync_bellatrix"),
		)...,
	)
}
func TestHarnessForkChoice(t *testing.T) {
	beacontest.Execute(
		append(
			defaultHarnessOpts(harnessConfig{t: t, v: clparams.BellatrixVersion, forkmode: 1}),
			beacontest.WithTestFromFs(Harnesses, "fork_choice"),
		)...,
	)
}
