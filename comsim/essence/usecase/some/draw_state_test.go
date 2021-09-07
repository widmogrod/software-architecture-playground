package some

import (
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
      "OutputPath": "$.__vars__.input.var_value",
      "Type": "Pass"
    },
    "Start1": {
      "Next": "Ok4",
      "ResultPath": "$.__vars__.input.var_value",
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
      "ResultPath": "$.__vars__.input.var_value",
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
        "ok.$": "$.__vars__.input.var_value.Id"
      },
      "Type": "Pass"
    },
    "Start1": {
      "Next": "Ok4",
      "ResultPath": "$.__vars__.input.var_value",
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
              "Variable": "$.__vars__.input.var_value.Id"
            },
            {
              "And": [
                {
                  "StringEquals": "Prometheus",
                  "Variable": "$.__vars__.input.var_value.Name"
                },
                {
                  "BooleanEquals": true,
                  "Variable": "$.__vars__.input.var_value.Alive"
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
      "ResultPath": "$.__vars__.input.var_value",
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
              "BooleanEqualsPath": "$.__vars__.input.var_value.Id2",
              "Variable": "$.__vars__.input.var_value.Id"
            },
            {
              "NumericEqualsPath": "$.__vars__.input.var_value.Id2",
              "Variable": "$.__vars__.input.var_value.Id"
            },
            {
              "StringEqualsPath": "$.__vars__.input.var_value.Id2",
              "Variable": "$.__vars__.input.var_value.Id"
            },
            {
              "TimestampEqualsPath": "$.__vars__.input.var_value.Id2",
              "Variable": "$.__vars__.input.var_value.Id"
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
      "ResultPath": "$.__vars__.input.var_value",
      "Type": "Pass"
    }
  }
}`,
		},
		"returns nested chose": {
			workflow: WorkparToWorkflow([]byte(`flow HelloWorld(input) {
	if eq(input.Id, 7) {		
	   if eq(input.Alive, true) {		
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
}`)),
			stateMachine: `{
  "Comment": "flow (input)",
  "StartAt": "Start1",
  "States": {
    "Choose10": {
      "Choices": [
        {
          "Next": "Ok13",
          "NumericEquals": 10,
          "Variable": "$.__vars__.input.var_value.Id"
        }
      ],
      "Default": "Err10",
      "Type": "Choice"
    },
    "Choose13": {
      "Choices": [
        {
          "Next": "Choose19",
          "NumericEquals": 7,
          "Variable": "$.__vars__.input.var_value.Id"
        }
      ],
      "Default": "Choose10",
      "Type": "Choice"
    },
    "Choose19": {
      "Choices": [
        {
          "BooleanEquals": true,
          "Next": "Ok22",
          "Variable": "$.__vars__.input.var_value.Alive"
        }
      ],
      "Default": "Err19",
      "Type": "Choice"
    },
    "Err10": {
      "End": true,
      "Parameters": {
        "ok2": false
      },
      "Type": "Pass"
    },
    "Err19": {
      "End": true,
      "Parameters": {
        "ok1": false
      },
      "Type": "Pass"
    },
    "Ok13": {
      "End": true,
      "Parameters": {
        "ok2": true
      },
      "Type": "Pass"
    },
    "Ok22": {
      "End": true,
      "Parameters": {
        "ok1": true
      },
      "Type": "Pass"
    },
    "Start1": {
      "Next": "Choose13",
      "ResultPath": "$.__vars__.input.var_value",
      "Type": "Pass"
    }
  }
}`,
		},
		"returns assigment": {
			workflow: WorkparToWorkflow([]byte(`flow HelloWorld(input) {
	a = EchoChamber({"Name": input.Name})
	b = EchoChamber({"Name": input.Name, "Other": a.body})
	return(b)
}`)),
			stateMachine: `{
  "Comment": "flow (input)",
  "StartAt": "Start1",
  "States": {
    "Assign4": {
      "Next": "Assign8",
      "Parameters": {
        "FunctionName": "arn:aws:lambda:eu-west-1:483648412454:function:EchoChamber",
        "Payload": {
          "Name.$": "$.__vars__.input.var_value.Name"
        }
      },
      "Resource": "arn:aws:states:::lambda:invoke",
      "ResultPath": "$.__vars__.a",
      "ResultSelector": {
        "var_value.$": "$.Payload"
      },
      "Retry": [
        {
          "BackoffRate": 2,
          "ErrorEquals": [
            "Lambda.ServiceException",
            "Lambda.AWSLambdaException",
            "Lambda.SdkClientException"
          ],
          "IntervalSeconds": 2,
          "MaxAttempts": 1
        }
      ],
      "Type": "Task"
    },
    "Assign8": {
      "Next": "Ok12",
      "Parameters": {
        "FunctionName": "arn:aws:lambda:eu-west-1:483648412454:function:EchoChamber",
        "Payload": {
          "Name.$": "$.__vars__.input.var_value.Name",
          "Other.$": "$.__vars__.a.var_value.body"
        }
      },
      "Resource": "arn:aws:states:::lambda:invoke",
      "ResultPath": "$.__vars__.b",
      "ResultSelector": {
        "var_value.$": "$.Payload"
      },
      "Retry": [
        {
          "BackoffRate": 2,
          "ErrorEquals": [
            "Lambda.ServiceException",
            "Lambda.AWSLambdaException",
            "Lambda.SdkClientException"
          ],
          "IntervalSeconds": 2,
          "MaxAttempts": 1
        }
      ],
      "Type": "Task"
    },
    "Ok12": {
      "End": true,
      "OutputPath": "$.__vars__.b.var_value",
      "Type": "Pass"
    },
    "Start1": {
      "Next": "Assign4",
      "ResultPath": "$.__vars__.input.var_value",
      "Type": "Pass"
    }
  }
}`,
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			result, err := WorkflowToAWSStateMachine(uc.workflow)
			t.Log(result)
			assert.NoError(t, err)
			assert.JSONEq(t, uc.stateMachine, result)
		})
	}
}
