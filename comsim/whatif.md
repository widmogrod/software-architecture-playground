// Model representing logical relations between concepts
// model {

// Aggregate restricts relation to three
aggregate {
    Question {
        Id
        Answers Answer[] // this creates reference
        Content String  @delegate(["question_content_validator"])
        UserId  UserId
        Moderation Moderation {

        }
    }

    extend Question with QuestionModerated {

    }

    Answer {
        Id
        Content String
    }
}


// Stable business events, independent from physical implementation;
// higher-order events, not something like question-{created|updated|deleted} that can be generated automatically
events {
    QuestionNeedsAnswer {
        EventId     Uuid
        QuestionId  model.Question.Id
    }
    Question {

    }
}

emitters {
    Webhook(events.*)
    QuestionNeedsAnswer {


    }
}

// Express what queries will be issue in the system
queries {
    QuestionsToAnswer {
        Input {
        }
        Output {
            Result []model.Question with {
                Answer
            }
        }
    }
}

// Operations that will change data
mutations {
    CreateQuestion {
        Input {
            mode.Question,
        }
        Output {
            mode.Question,
        }
    }

    CreateQuestion2 {
        Input {
            Content      mode.Question.Content,
            UserIdentity runtime.Auth.Identity
        }
        Output {
            mode.Question,
        }
    }

    UpdateQuestionContent {
        Input {
            model.Question.Id
            runtime.Auth.Identity
        } @delegate(["can_modify_question"])
    }
}

checks {
    can_modify_question {
        Input runtime.Auth.Identity
    }
    question_content_validator {
        Input String
        Output {
            Validation tagged {
                ToShort {}
                ToLong {}
            }
        }
    }
}

runtime {
    Auth {
        Identity {
            Id      Int
            Market  String
        }
    }

}