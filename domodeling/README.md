# Domain Modeling in Golang type system

- State management
- Restricting programs

- [Parse, don't validate](https://lexi-lambda.github.io/blog/2019/11/05/parse-don-t-validate/)
- Names are not

> Parsing preserve information, validation don't.

--- 
> taxonomies are useful for documenting a domain of interest, but not necessarily helpful for modeling it
> Primarily, types distinguish functional differences between values. A value of type NonEmpty a is functionally distinct from a value of type [a], since it is fundamentally structurally different and permits additional operations. In this sense, types are structural; they describe what values are in the internal world of the programming language.
> Secondarily, we sometimes use types to help ourselves avoid making logical mistakes. We might use separate Distance and Duration types to avoid accidentally doing something nonsensical like adding them together, even though they’re both representationally real numbers.

https://lexi-lambda.github.io/blog/2020/11/01/names-are-not-type-safety/

> type system and checking is just a tool

We shoudn't write code to limit ourself too much

---

Re use types across modules?
- To specific `type Some string` 
- GraphQL federation

// DATA in

POST /some-action
{"some":"input"}

// Framework handles
rq := ParseHTTP()
shape := ParseJSON(rq.Body)

// Essence

// value objects, available to other sub-domains? - or how to compose subdomains?
domain_tokens := Money | Dimension | 
domain_object := { primitive + domain_tokens + domain_object} 

dom:=ParseDomain(shape) DomObj {
    DomObj.<field> = ParseValueObject(shape.value)
}

int:=InterpretDomainA() {
    HandleCreateX() {
        // create other domain objects
        // invoke other interpreters? or execute other domain objects
        // expand or reduce domain execution
        // communicate with storage
        // - builds SQL (express change in domain language)
        // - maps back results
        // ---- storage layer like SQL parse everything once again and valdates integrity and constraings
        // 
        // Where is integrity constrains check in this model?
    }
}

res:=ExecuteDomainA(dom,int) 

// Framework

DecodeJSON()
SendHTTP()

// DATA out

HTTP 200
{success:}



------

   question
   - id : question_id
   - author: user_id
   - content: text

   answer
   - id: answer_id
   - author: user_id
   - question: question_id  
   - content: text

   comment
   - id:comment_id
   - author: user_id
   - content: text
   - created|updated|deleted : date

   ref_has_comment
   - object_ref: question_id | answer_id
   - comment_ref: comment_id

-- view requirements

   question_feed_view: {filter_by: [user_id, number_of_answers], sort_by: [created_at]}
   - question_ref: question_id
   - question_author_ref: user_id
   - question_created: date
   - number_of_answers: int
   - number_of_question_comments: int
   - number_of_answer_comments: int

   how to keep it with sync?
   - optimistic UI
   
   how to make it fast, without overcomplicated state management?
   - materialise views - requires batch updates - eventual consistency
   - projections re-building with streaming, CDCs, event sourcing, outbox pattern + events - eventual consistency
   - HTAP databases - eventual consistency
   
   how to make it?
   - joins - via SQL
   - aggregations on API Gateway like GraphQL for simplere 1:1 aggregates can be faster
   - backend for frontend with their own state management - write through back-end for frontend?
  
 
## how to structure team ownership and independence in monorepo?

- soft delete is interesting decision showing tight coupling? 
  do we really need to validate whenever question exits? 
  imagine that we now try to separate part of the system and `comment.add` requires do remote API call
  Thanks to approaching external comm and internal comm in the same way, programming becomes simplere
- 

/src
    /source-of-truth
        /<question>
            create(input) {
                INSERT INTO question(content,author) 
                     VALUES (:input.content,:input.userId)
            }
            get(input) {
                SELECT * FROM question WHERE id = :input.questionId AND deleted IS NOT NULL
            }
        /<answer>
        /<comment>
            @depends(question, answer, user) 
            add(input) {
                @user := user.get(input.userId)
                switch input.type {
                case "question":
                    @question := question.get(@data.questionId)     
                    self.private-add({type: question, objectId: @question.id, userId: @user.id, comment: input.comment.sanitize()})
                case "answer":
                    @answer := question.get(@data.answerId)     
                    self.private-add({type: answer, objectId: @answer.id, userId: @user.id, comment: input.comment.sanitize()})
            }
            private-add(input) {
                BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE READ ONLY DEFERRABLE
                @cid := INSERT INTO comment(author,content) 
                     VALUES (:input.userId, :input.content)
                     RETURNING id
                INSERT INTO ref_has_comment(object_ref, comment_ref)
                     VALUES (@id, :input.type, :input.objectId)
                COMMIT
            }
            add-workflow(input) {
                INSERT INTO workflow (name, data)
                     VALUES ('add-comment', input:data)
            }
            execute-workflow(input) {
                @data = SELECT data FROM workflow WHERE id := input.id
                if @data.status != "new"; return
                @res = self.add(@data)
                UPDATE workflow SET status="ok", res=@res WHERE id = @data.id
            }
            @cron("* * * * *")
            @stream("cdc-workflow")
            process(input) {
                switch input {
                    Schedule() => 
                        @workflows = SELECT id FROM workflow WHERE status = "new" LIMIT 10
                        @workflows.map( x => self.execute-workflow(x.id))
                    Change(inserts) => 
                        inserts.map( x => self.execute-workflow(id))
                }
            }
    /projections
        /question_feed_view


let's add as a new requirements points or other business constrain

----

Code generation at B for μservice is interesting productivity gain, but one step further is 
how can we then remove what engineers need to generate and maintain nad upload only business logic
and runtime we be updated and managed by other team?