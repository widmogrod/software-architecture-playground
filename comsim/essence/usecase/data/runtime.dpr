predicate
 =     Eq (path, value)
 | Exists (path)
 |    And (predicate, predicate)
 |     Or (predicate, predicate)
;

workflow
 =   Activity (AID, activityT)
 | Transition (workflow, workflow)
;
 // | Transition {from: workflow, to: workflow}

activityT
 = start
 | end(endT)
 | choose (predicate, Activity, Activity)
 | reshape(reshapeT)
 | invocation (fid)
;

endT
 = ok
 | err
;

reshapeT
 = rpath [lit]
 | rlist [path]
 | rdict [(path, reshapeT)]
;

//data
// = dlit (lit)
// | dlist [data]
// | ddict [(lit, data)]
//;