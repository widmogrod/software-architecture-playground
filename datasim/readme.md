# Data playground
Imagine that you join new company and in your first day of work you're able to release production quality backend API
that can be used by millions of users, without additional effort, like lengthy design sessions, configuring monitoring and alerting, etc.
Newly released backend API can be use by any of the company's teams by simply installing via package manager SDK to a service.

Sound crazy? Definitely not, it's reachable, we have to find answer to question what is required to reach such vision?
- Automation without doubt, with strong technical opinions and understanding of company needs.
- What means production quality?
  - Observability
  - Backups
  - Security
  - Cost visibility
  - ...
  - (well-architected)

- lifecyckle of databases, moving data, unbound growth, indexing, sharing, 

Risks
- ORM like
- hard to change

> Idea. Imagine that everytign could be event "click" in wisard, you define data schema, click deploy,
> you see that SDKs for X platforms were generated, and which platform/product use them, and which version
> Imagine that after X days you change schema, add fields, add validation, also through UI; because system is deeply integrated
> because it generates and manage SDKs, and tracks SKDs usage version, and there is automation to automatically issue Pull-Request to projects to update sdk to new version
> because solution is designed as multi-tenant you can promote change slowly across tenants
> because solution validates schema for backward compatibility, change in data reduce breaking change; and change can be roll back
> because solution has automatic process that manage common lifecycles like; move data, shard data, scale down tenant; rollback or promote change,...
> 
> Even if changes created by UI are considered as a "bad practice" in today's practice, because it's not reproducible and versioned in code
> Still with so much automation and possibility to generate new service that fix issues with previous one in just "one day" can be considered as paridigm-shift 
> It's cheaper to rebuild than to fix or improve it; Maybe even it can become Jevons paradox.

## What I would like to learn?
- I would also like to know what are areas to compensate by in-house investment,
- I would like to know performance to cost-of-running to cost-of-operating 
- I would like to know breaking points, hard-limits,...

## What hypothesis I would test?
- It's worth to use X-approach than Y-approach, and compensate differences by investing in developing in-house solutions
- Back-End as a Service is still sensible option in 2022 as in-house investment (build vs buy)

## Methodology
- Define conceptual model of data that solution will work out-the-box
- Define how each conceptual model can be projected to logical/physical model.
- Define access patterns that collect metrics and help draw conclusions
  - Create N entities with K attributes
  - Update N entities with K attributes
  - X random reads from datastore of N entities and K attributes
  - ...

For each "logical projection" (implementation) measure and understand
- Understand how it behaves under unbound growth like:
  - how many single-key-lookups per second it can handle      (experiment-id, "single-key-lopups", [time-diff=123])
    - how it changes with growth of storage                   (experiment-id, "single-key-lopups", ["table-items", 10000000])
    - how it changes with number of attributes                (experiment-id, "single-key-lopups", ["no_attr_primary" = 20])
  - how it performs when number of sub-entities grows 


## Data modeling, why, technics, keywords
Ontology
- Ontology engineering
  - Data modeling (Conceptual Model; Logical Model; Physical Model)
  - Information systems
  - Data as a resource

"Start from hardest thing to get right first"

There are many data models, and each of them has unique properties.
- Storage
- Representation
- Manipulation

SCHEMA
- Semantic schema.
- Validating Schema
- Emergent schema.

IDENTITY
- Persistent identifiers
- External identity links
- Datatypes
- Lexicalisation
- Existential nodes.

CONTEXT

- Document 
- Relational (tabular format)
- Graph (triples)
  - Directed edge-labelled graphs.
  - Heterogeneous graphs.
  - Property graphs.
  - Graph dataset.
- Key-Value
  - FoundationDB
  - BigTable
  - Hazelcast, 
  - DynamoDB

- Entity-attribute-Value
- Temporal
- Spartial

- Inverted Index
- Triplestore (subject-predicate-object)
- Named Ggraph

---

- Impedance Mismatch - ORM - mapping from objects to:

Ne - # entities
Nf - # fields

## assumption
- each field has same fixed size
- "row" is one line in a file
  - disk sector has size 4KiB //kibibyte (1024)
  - block size is 4kbs
  - some file systems require block size to be occupide by one file, which may result in waste space when many small files is included
  - which means that one disk read more that it needs

## (entity) => (id, fields...)
// all fields in one row
cost(select1 entity) = 1 * row                     =  1 operations
cost(join entities)  = cost(select1 entity) * Ne   = Ne operations

## (entity, field) => (id, value:type)
// all fields in separate rows
cost(select1 entity) = Nf * row  = Nf operations   = Nf operations
cost(join entities)  = cost(select1 entity) * Ne   = Ne * Nf operations

---

(entity) => (id, fields...)

Question(id, content, created, author_id,...)
Author(id, name,...)
Answer(id, question_id, author_id, content)

vs

(entity, field) => (qid, value:type)

QuestionContent(id, content)
QuestionCreated(id, created)
QuestionAuthor(id, author_id)
AuthorName(id, name)
Author

vs

(entity, type) => (id, field, value)

QuestionString(id, field, value)
QuestionInt(id, field, value)
AuthorString(id, field, value)
AuthorInt(id, field, value)

StringAuthor(id, value)
StringQuestions(id, value)

StringContent(id, question, author, ...)
IntAuthor(id, question, answer, ...)


vs

(entity) => (id, value)

Question(id, {
  "content": "",
  "created": "",
  "author": {
    "id": "",
    "name": "",
  },
  "answers": [
    {
      "id": "",
      "content": "",
    },
  ],
})

How do we query this data?

SQL++

vs

map(str,str)




---
References
- https://www.brianlikespostgres.com/cost-of-a-join.html
- https://blog.codinghorror.com/object-relational-mapping-is-the-vietnam-of-computer-science/
