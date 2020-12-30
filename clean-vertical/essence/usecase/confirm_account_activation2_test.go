package usecase

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
)

func TestAsyncHandleConfirmAccountActivation(t *testing.T) {
	token := "activation token"
	ctx := dispatch.Background()
	res := AsyncHandleConfirmAccountActivation(ctx, ConfirmAccountActivation{
		ActivationToken: token,
	})

	assert.NotEmpty(t, res.InvocationID)

	// TODO Improve registration of workflows
	//
	//  flow := dispatch.NewFlowRegistry()
	//  flow.Register('name', AsyncHandleConfirmAccountActivation)
	//
	//  wf := flow.RetrieveFlow('name')
	//
	//  assert(dispatch.ToPlantText(wf) === '[*] -->')
	//
	//  res := flow.Invoke(ctx, ConfirmAccountActivation{
	//		ActivationToken: token,
	//  })
	//
	//  runtime := dispatch.NewInMemoryFlowRuntime()
	//  runtime.Start(flow)
	//
	//  assert(runtime.GetStatus(res.InvocationID) == runtime.COMPLETE)
	//  assert(runtime.GetResult(res.InvocationID) == {})

	wf := dispatch.RetrieveFlow(res.InvocationID)

	assert.Equal(t,
		dispatch.ToPlantText(wf),
		`[*] --> MarkAccountActivationTokenAsUse: run 
MarkAccountActivationTokenAsUse --> ResultOfMarkingAccountActivationTokenAsUsed  
ResultOfMarkingAccountActivationTokenAsUsed --> if_ResultOfMarkingAccountActivationTokenAsUsed  
if_ResultOfMarkingAccountActivationTokenAsUsed --> [*]: then 
if_ResultOfMarkingAccountActivationTokenAsUsed --> GenerateSessionToken: else 
GenerateSessionToken --> ResultOfGeneratingSessionToken  
ResultOfGeneratingSessionToken --> [*]  
`)

	t.Log("TODO: pooling for result")
}
