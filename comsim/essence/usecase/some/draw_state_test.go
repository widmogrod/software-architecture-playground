package some

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/usecase/data"
	"testing"
)

func TestDraw(t *testing.T) {
	useCases := map[string]struct {
		workflow     data.Workflow
		stateMachine string
	}{
		"returns input": {
			workflow: WorkparToWorkflow([]byte(`flow HelloWorld(input) {
		return(input)
}`)),
			stateMachine: `{
  "Comment": "flow (input)",
  "StartAt": "Start1",
  "States": {
    "Ok4": {
      "End": true,
      "OutputPath": "$.input",
      "Type": "Pass"
    },
    "Start1": {
      "Next": "Ok4",
      "ResultPath": "$.input",
      "Type": "Pass"
    }
  }
}`,
		},
		"returns literal value": {
			workflow: WorkparToWorkflow([]byte(`flow HelloWorld(input) {
		return({"ok": 7})
}`)),
			stateMachine: `{
  "Comment": "flow (input)",
  "StartAt": "Start1",
  "States": {
    "Ok4": {
      "End": true,
      "Parameters": {
        "ok": 7
      },
      "Type": "Pass"
    },
    "Start1": {
      "Next": "Ok4",
      "ResultPath": "$.input",
      "Type": "Pass"
    }
  }
}`,
		},
		"returns dynamic mapping": {
			workflow: WorkparToWorkflow([]byte(`flow HelloWorld(input) {
		return({"ok": input.Id})
}`)),
			stateMachine: `{
  "Comment": "flow (input)",
  "StartAt": "Start1",
  "States": {
    "Ok4": {
      "End": true,
      "Parameters": {
        "ok.$": "$.input.Id"
      },
      "Type": "Pass"
    },
    "Start1": {
      "Next": "Ok4",
      "ResultPath": "$.input",
      "Type": "Pass"
    }
  }
}`,
		},
		"returns chose": {
			workflow: WorkparToWorkflow([]byte(`flow HelloWorld(input) {
	if eq(input.Id, 7) {		
		return({"ok": true})
	} else {
		fail({"ok": false})
	}
}`)),
			stateMachine: `{
  "Comment": "flow (input)",
  "StartAt": "Start1",
  "States": {
    "Choose4": {
      "Choices": [
        {
          "Next": "Ok7",
          "NumericEquals": 7,
          "Variable": "$.input.Id"
        }
      ],
      "Default": "Err10",
      "Type": "Choice"
    },
    "Err10": {
      "End": true,
      "Parameters": {
        "ok": false
      },
      "Type": "Pass"
    },
    "Ok7": {
      "End": true,
      "Parameters": {
        "ok": true
      },
      "Type": "Pass"
    },
    "Start1": {
      "Next": "Choose4",
      "ResultPath": "$.input",
      "Type": "Pass"
    }
  }
}`,
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			result, err := WorkflowToAWSStateMachine(uc.workflow)
			fmt.Println(result)
			assert.NoError(t, err)
			assert.JSONEq(t, uc.stateMachine, result)
		})
	}
}
