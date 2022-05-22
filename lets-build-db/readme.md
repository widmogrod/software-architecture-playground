# Introduction

Let's build db, is simple project when I learn basics of databases internals on example on building very simple database
engines

## Example `1_naive.go`

This database is build around of concept storing `<key, values>`. 
When someone would want to a record like this:

```json
{
  "id": "666",
  "content": "some content",
  "createdAt": "2022-05-20"
}
```

Then it would need to be mapped to key-value as:

```
["record:666:content", "some content"]
["record:666:createdAt", "2022-05-20"]
```

- Database provides only two operations - Set and Get.
- Database is build around concept of append log. 
  - Which means that any Set operations append values to the end of the log
    - Key-value operations can be grouped. 
      Grouping must be sorted by key. 
      Such sorting is responsibility of client code, not DB code.
  - Which means that any Get operations to retrieve value under key need to scan append-log from **end** to beginning

- Any operation like "find" attributes that match criteria are not supported.
  - Such operation can be build by reading append-only-log and constructing indices optimised for search operation

- Append-log should be separated in segments that can be replicated 
  - Segment when flush should contain hash of previous segment, to build a chain

- Such model acts also last-win-writes. 
  There is no semantic that would prevent such situation. 
  Last-win writes doesn't mean that we loose data, previous values stay in the log.
- Because of this to reduce space, process of compacting log is necessary
- How to ensure semantic that prevents race conditions?

- How to scale whole system on many nodes?
  

