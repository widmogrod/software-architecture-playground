predicate
 =     Eq {path: path, value: value}
 | Exists {path: path}
 |    And (predicate, predicate)
 |     Or (predicate, predicate)
;

workflow
 =   Activity {id: AID, activity: activityT}
 | Transition {from: workflow, to: workflow}
;

activityT
 =      Start
 |        End = end
 |     Choose {if: predicate, then: workflow, else: workflow}
 |    Reshape = reshape
 | Invocation (fid)
;

end
 = Ok
 | Err
;

reshape
 = Select {Path: path}
 |  ReMap [{Key: path, Value: path}]
;



