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
		"returns chose equal to scalar": {
			workflow: WorkparToWorkflow([]byte(`flow HelloWorld(input) {
	if or(eq(input.Id, 0.3), and(eq(input.Name, "Prometheus"), eq(input.Alive, true))) {		
		return({"ok": true})
	} else {
		fail({"ok": false})
	}
}`)),
			stateMachine: `{
  "Comment": "flow (input)",
  "StartAt": "Start1",
  "States": {
    "Choose7": {
      "Choices": [
        {
          "Next": "Ok10",
          "Or": [
            {
              "NumericEquals": 0.3,
              "Variable": "$.input.Id"
            },
            {
              "And": [
                {
                  "StringEquals": "Prometheus",
                  "Variable": "$.input.Name"
                },
                {
                  "BooleanEquals": true,
                  "Variable": "$.input.Alive"
                }
              ]
            }
          ]
        }
      ],
      "Default": "Err7",
      "Type": "Choice"
    },
    "Err7": {
      "End": true,
      "Parameters": {
        "ok": false
      },
      "Type": "Pass"
    },
    "Ok10": {
      "End": true,
      "Parameters": {
        "ok": true
      },
      "Type": "Pass"
    },
    "Start1": {
      "Next": "Choose7",
      "ResultPath": "$.input",
      "Type": "Pass"
    }
  }
}`,

		},
		"returns choose dynamic values": {
			workflow: WorkparToWorkflow([]byte(`flow HelloWorld(input) {
	if eq(input.Id, input.Id2) {		
		return({"ok": true})
	} else {
		fail({"ok": false})
	}
}`)),
			stateMachine: `{
  "Comment": "flow (input)",
  "StartAt": "Start1",
  "States": {
    "Choose7": {
      "Choices": [
        {
          "Next": "Ok10",
          "Or": [
            {
              "BooleanEqualsPath": "$.input.Id2",
              "Variable": "$.input.Id"
            },
            {
              "NumericEqualsPath": "$.input.Id2",
              "Variable": "$.input.Id"
            },
            {
              "StringEqualsPath": "$.input.Id2",
              "Variable": "$.input.Id"
            },
            {
              "TimestampEqualsPath": "$.input.Id2",
              "Variable": "$.input.Id"
            }
          ]
        }
      ],
      "Default": "Err7",
      "Type": "Choice"
    },
    "Err7": {
      "End": true,
      "Parameters": {
        "ok": false
      },
      "Type": "Pass"
    },
    "Ok10": {
      "End": true,
      "Parameters": {
        "ok": true
      },
      "Type": "Pass"
    },
    "Start1": {
      "Next": "Choose7",
      "ResultPath": "$.input",
      "Type": "Pass"
    }
  }
}`,

		},
		"returns nested chose": {
			workflow: WorkparToWorkflow([]byte(`flow HelloWorld(input) {
	if eq(input.Id, 7) {		
		if eq(input.Id, 9) {		
			return({"ok1": true})
		} else {
			fail({"ok1": false})
		}
	} else {
		if eq(input.Id, 10) {		
			return({"ok2": true})
		} else {
			fail({"ok2": false})
		}
	}
	fail({"ok3": false})
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
