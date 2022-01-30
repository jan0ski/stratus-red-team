package runner

import (
	"testing"

	statemocks "github.com/datadog/stratus-red-team/internal/state/mocks"
	"github.com/datadog/stratus-red-team/pkg/stratus"
	"github.com/datadog/stratus-red-team/pkg/stratus/runner/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRunnerWarmUp(t *testing.T) {

	type RunnerWarmupTestScenario struct {
		Name                  string
		Technique             *stratus.AttackTechnique
		ShouldForce           bool
		InitialTechniqueState stratus.AttackTechniqueState
		TerraformOutputs      map[string]string
		PersistedOutputs      map[string]string
		// results
		CheckExpectations func(t *testing.T, terraform *mocks.ConfigManager, state *statemocks.StateManager, outputs map[string]string, err error)
	}

	var scenario = []RunnerWarmupTestScenario{
		{
			Name:                  "Warming up a technique without prerequisite Terraform code",
			Technique:             &stratus.AttackTechnique{ID: "foo"},
			InitialTechniqueState: stratus.AttackTechniqueStatusCold,
			PersistedOutputs:      map[string]string{"myoutput": "foo"},
			CheckExpectations: func(t *testing.T, terraform *mocks.ConfigManager, state *statemocks.StateManager, outputs map[string]string, err error) {
				terraform.AssertNotCalled(t, "InitAndApply")
				state.AssertNotCalled(t, "ExtractTechnique")
				assert.Nil(t, err)

				// No prerequisite Terraform code implies there cannot be any output
				assert.Len(t, outputs, 0)
			},
		},
		{
			Name:                  "Warming up a COLD technique",
			Technique:             &stratus.AttackTechnique{ID: "foo", PrerequisitesTerraformCode: []byte("foo")},
			InitialTechniqueState: stratus.AttackTechniqueStatusCold,
			TerraformOutputs:      map[string]string{"myoutput": "new"},
			CheckExpectations: func(t *testing.T, terraform *mocks.ConfigManager, state *statemocks.StateManager, outputs map[string]string, err error) {
				state.AssertCalled(t, "ExtractTechnique")
				terraform.AssertCalled(t, "InitAndApply")
				state.AssertCalled(t, "WriteOutputs", map[string]string{"myoutput": "new"})
				state.AssertCalled(t, "SetTechniqueState", stratus.AttackTechniqueState(stratus.AttackTechniqueStatusWarm))

				assert.Nil(t, err)
				assert.Len(t, outputs, 1)
				assert.Equal(t, "new", outputs["myoutput"])
			},
		},
		{
			Name:                  "Warming up a WARM technique without force flag",
			Technique:             &stratus.AttackTechnique{ID: "foo", PrerequisitesTerraformCode: []byte("bar")},
			InitialTechniqueState: stratus.AttackTechniqueStatusWarm,
			PersistedOutputs:      map[string]string{"myoutput": "new"},
			CheckExpectations: func(t *testing.T, terraform *mocks.ConfigManager, state *statemocks.StateManager, outputs map[string]string, err error) {
				terraform.AssertNotCalled(t, "InitAndApply")
				assert.Nil(t, err)
				assert.Len(t, outputs, 1)
				assert.Equal(t, "new", outputs["myoutput"])
			},
		},
		{
			Name:                  "Warming up a WARM technique with force flag",
			Technique:             &stratus.AttackTechnique{ID: "foo", PrerequisitesTerraformCode: []byte("bar")},
			ShouldForce:           true,
			InitialTechniqueState: stratus.AttackTechniqueStatusWarm,
			TerraformOutputs:      map[string]string{"myoutput": "old"},
			CheckExpectations: func(t *testing.T, terraform *mocks.ConfigManager, state *statemocks.StateManager, outputs map[string]string, err error) {
				terraform.AssertCalled(t, "InitAndApply")
				assert.Nil(t, err)
				assert.Len(t, outputs, 1)
				assert.Equal(t, "old", outputs["myoutput"])
			},
		},
		{
			Name:                  "Warming up a DETONATED technique",
			Technique:             &stratus.AttackTechnique{ID: "foo", PrerequisitesTerraformCode: []byte("bar")},
			InitialTechniqueState: stratus.AttackTechniqueStatusDetonated,
			PersistedOutputs:      map[string]string{"myoutput": "old"},
			CheckExpectations: func(t *testing.T, terraform *mocks.ConfigManager, state *statemocks.StateManager, outputs map[string]string, err error) {
				terraform.AssertNotCalled(t, "InitAndApply")
				assert.Nil(t, err)
				assert.Len(t, outputs, 1)
				assert.Equal(t, "old", outputs["myoutput"])
			},
		},
	}

	for i := range scenario {
		state := new(statemocks.StateManager)
		terraform := new(mocks.ConfigManager)

		state.On("GetRootDirectory").Return("/root")
		state.On("ExtractTechnique").Return(nil)
		state.On("GetTechniqueState", mock.Anything).Return(scenario[i].InitialTechniqueState, nil)
		state.On("GetOutputs").Return(scenario[i].PersistedOutputs, nil)
		terraform.On("InitAndApply", mock.Anything).Return(scenario[i].TerraformOutputs, nil)
		state.On("WriteOutputs", mock.Anything).Return(nil)
		state.On("SetTechniqueState", mock.Anything).Return(nil)

		runner := Runner{
			Technique:     scenario[i].Technique,
			ShouldForce:   scenario[i].ShouldForce,
			ConfigManager: terraform,
			StateManager:  state,
		}
		runner.initialize()
		outputs, err := runner.WarmUp()
		t.Run(scenario[i].Name, func(t *testing.T) { scenario[i].CheckExpectations(t, terraform, state, outputs, err) })
	}
}

func TestRunnerDetonate(t *testing.T) {

	type TestDetonationScenario struct {
		Name                            string
		TechniqueState                  stratus.AttackTechniqueState
		IsIdempotent                    bool
		Force                           bool
		ExpectDetonated                 bool
		ExpectWarmedUp                  bool
		ExpectError                     bool
		ExpectedStateChangedToDetonated bool
	}

	scenario := []TestDetonationScenario{
		{
			Name:                            "DetonateWarmIdempotentAttackTechnique",
			TechniqueState:                  stratus.AttackTechniqueStatusWarm,
			IsIdempotent:                    true,
			Force:                           false,
			ExpectWarmedUp:                  false,
			ExpectDetonated:                 true,
			ExpectError:                     false,
			ExpectedStateChangedToDetonated: true,
		},
		{
			Name:                            "DetonateWarmNonIdempotentAttackTechnique",
			TechniqueState:                  stratus.AttackTechniqueStatusWarm,
			IsIdempotent:                    false,
			Force:                           false,
			ExpectWarmedUp:                  false,
			ExpectDetonated:                 true,
			ExpectError:                     false,
			ExpectedStateChangedToDetonated: true,
		},
		{
			Name:                            "DetonateDetonatedIdempotentAttackTechnique",
			TechniqueState:                  stratus.AttackTechniqueStatusDetonated,
			IsIdempotent:                    true,
			Force:                           false,
			ExpectWarmedUp:                  false,
			ExpectDetonated:                 true,
			ExpectError:                     false,
			ExpectedStateChangedToDetonated: true,
		},
		{
			Name:                            "DetonateDetonatedNonIdempotentAttackTechnique",
			TechniqueState:                  stratus.AttackTechniqueStatusDetonated,
			IsIdempotent:                    false,
			Force:                           false,
			ExpectWarmedUp:                  false,
			ExpectDetonated:                 false,
			ExpectError:                     true,
			ExpectedStateChangedToDetonated: false,
		},
		{
			Name:                            "DetonateDetonatedNonIdempotentAttackTechniqueWithForceFlag",
			TechniqueState:                  stratus.AttackTechniqueStatusDetonated,
			IsIdempotent:                    false,
			Force:                           true,
			ExpectWarmedUp:                  false,
			ExpectDetonated:                 true,
			ExpectError:                     false,
			ExpectedStateChangedToDetonated: true,
		},
	}

	for i := range scenario {
		t.Run(scenario[i].Name, func(t *testing.T) {
			state := new(statemocks.StateManager)
			terraform := new(mocks.ConfigManager)

			state.On("GetRootDirectory").Return("/root")
			state.On("ExtractTechnique").Return(nil)
			state.On("GetTechniqueState", mock.Anything).Return(scenario[i].TechniqueState, nil)
			terraform.On("InitAndApply", mock.Anything).Return(map[string]string{}, nil)
			state.On("WriteOutputs", mock.Anything).Return(nil)
			state.On("GetOutputs").Return(map[string]string{}, nil)
			state.On("SetTechniqueState", mock.Anything).Return(nil)

			var wasDetonated = false
			runner := Runner{
				Technique: &stratus.AttackTechnique{
					ID: "sample-technique",
					Detonate: func(map[string]string) error {
						wasDetonated = true
						return nil
					},
					IsIdempotent: scenario[i].IsIdempotent,
				},
				ShouldForce:   scenario[i].Force,
				ConfigManager: terraform,
				StateManager:  state,
			}
			runner.initialize()
			err := runner.Detonate()

			if scenario[i].ExpectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			if scenario[i].ExpectWarmedUp {
				terraform.AssertCalled(t, "InitAndApply", mock.Anything)
			} else {
				terraform.AssertNotCalled(t, "InitAndApply", mock.Anything)
			}

			if scenario[i].ExpectDetonated {
				assert.True(t, wasDetonated)
			} else {
				assert.False(t, wasDetonated)
			}

			if scenario[i].ExpectedStateChangedToDetonated {
				state.AssertCalled(t, "SetTechniqueState", stratus.AttackTechniqueState(stratus.AttackTechniqueStatusDetonated))
			} else {
				state.AssertNotCalled(t, "SetTechniqueState", mock.Anything)
			}
		})
	}
}

func TestRunnerRevert(t *testing.T) {
	type TestRevertScenario struct {
		Name                        string
		TechniqueState              stratus.AttackTechniqueState
		Force                       bool
		ExpectDidCallRevertFunction bool
		ExpectDidChangeStateToWarm  bool
		ExpectError                 bool
	}
	scenario := []TestRevertScenario{
		{
			Name:                        "DetonatedTechniqueIsReverted",
			TechniqueState:              stratus.AttackTechniqueStatusDetonated,
			Force:                       false,
			ExpectDidCallRevertFunction: true,
			ExpectDidChangeStateToWarm:  true,
			ExpectError:                 false,
		},
		{
			Name:                        "WarmTechniqueIsNotReverted",
			TechniqueState:              stratus.AttackTechniqueStatusWarm,
			Force:                       false,
			ExpectDidCallRevertFunction: false,
			ExpectDidChangeStateToWarm:  false,
			ExpectError:                 true,
		},
		{
			Name:                        "WarmTechniqueIsRevertedWithForce",
			TechniqueState:              stratus.AttackTechniqueStatusWarm,
			Force:                       true,
			ExpectDidCallRevertFunction: true,
			ExpectDidChangeStateToWarm:  true,
			ExpectError:                 false,
		},
	}

	for i := range scenario {
		t.Run(scenario[i].Name, func(t *testing.T) {
			state := new(statemocks.StateManager)
			state.On("GetRootDirectory").Return("/root")
			state.On("ExtractTechnique").Return(nil)
			state.On("GetOutputs").Return(map[string]string{"foo": "bar"}, nil)
			state.On("GetTechniqueState", mock.Anything).Return(scenario[i].TechniqueState)
			state.On("SetTechniqueState", mock.Anything).Return(nil)

			var wasReverted = false
			runner := Runner{
				Technique: &stratus.AttackTechnique{
					ID:       "foo",
					Detonate: func(map[string]string) error { return nil },
					Revert: func(params map[string]string) error {
						wasReverted = true
						return nil
					},
				},
				ShouldForce:  scenario[i].Force,
				StateManager: state,
			}
			runner.initialize()

			err := runner.Revert()

			if scenario[i].ExpectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			if scenario[i].ExpectDidCallRevertFunction {
				assert.True(t, wasReverted)
			} else {
				assert.False(t, wasReverted)
			}

			if scenario[i].ExpectDidChangeStateToWarm {
				state.AssertCalled(t, "SetTechniqueState", stratus.AttackTechniqueState(stratus.AttackTechniqueStatusWarm))
			} else {
				state.AssertNotCalled(t, "SetTechniqueState", stratus.AttackTechniqueState(stratus.AttackTechniqueStatusWarm))
			}
		})
	}

}

func TestRunnerCleanup(t *testing.T) {
	type RunnerCleanupTestScenario struct {
		Name                  string
		Technique             *stratus.AttackTechnique
		ShouldForce           bool
		InitialTechniqueState stratus.AttackTechniqueState
		// results
		CheckExpectations func(t *testing.T, terraform *mocks.ConfigManager, state *statemocks.StateManager, err error)
	}

	var scenario = []RunnerCleanupTestScenario{
		{
			Name:                  "Cleaning up an already cold technique without force flag",
			Technique:             &stratus.AttackTechnique{ID: "foo", PrerequisitesTerraformCode: []byte("foo)")},
			InitialTechniqueState: stratus.AttackTechniqueStatusCold,
			CheckExpectations: func(t *testing.T, terraform *mocks.ConfigManager, state *statemocks.StateManager, err error) {
				assert.NotNil(t, err)
				terraform.AssertNotCalled(t, "Destroy")
				state.AssertNotCalled(t, "CleanupTechnique")
			},
		},
		{

			Name:                  "Cleaning up an already cold technique with force flag",
			Technique:             &stratus.AttackTechnique{ID: "foo", PrerequisitesTerraformCode: []byte("foo)")},
			InitialTechniqueState: stratus.AttackTechniqueStatusCold,
			ShouldForce:           true,
			CheckExpectations: func(t *testing.T, terraform *mocks.ConfigManager, state *statemocks.StateManager, err error) {
				assert.Nil(t, err)
				terraform.AssertCalled(t, "Destroy", mock.Anything)
				state.AssertCalled(t, "CleanupTechnique")
			},
		},
		{
			Name:                  "Cleaning up a WARM technique",
			Technique:             &stratus.AttackTechnique{ID: "foo", PrerequisitesTerraformCode: []byte("foo)")},
			InitialTechniqueState: stratus.AttackTechniqueStatusWarm,
			CheckExpectations: func(t *testing.T, terraform *mocks.ConfigManager, state *statemocks.StateManager, err error) {
				assert.Nil(t, err)
				terraform.AssertCalled(t, "Destroy", mock.Anything)
				state.AssertCalled(t, "CleanupTechnique")
				state.AssertCalled(t, "SetTechniqueState", stratus.AttackTechniqueState(stratus.AttackTechniqueStatusCold))
			},
		},
	}

	for i := range scenario {
		state := new(statemocks.StateManager)
		terraform := new(mocks.ConfigManager)

		state.On("GetRootDirectory").Return("/root")
		state.On("ExtractTechnique").Return(nil)
		state.On("GetTechniqueState", mock.Anything).Return(scenario[i].InitialTechniqueState, nil)
		state.On("SetTechniqueState", mock.Anything).Return(nil)
		state.On("CleanupTechnique").Return(nil)
		terraform.On("Destroy", mock.Anything).Return(nil)
		runner := Runner{
			Technique:     scenario[i].Technique,
			ShouldForce:   scenario[i].ShouldForce,
			ConfigManager: terraform,
			StateManager:  state,
		}
		runner.initialize()
		err := runner.CleanUp()
		t.Run(scenario[i].Name, func(t *testing.T) { scenario[i].CheckExpectations(t, terraform, state, err) })
	}
}
