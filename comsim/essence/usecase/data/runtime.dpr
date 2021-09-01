predicate
 =     Eq {path: path, value: any}
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
 |     Assign {Var: string, flow: workflow}
 |    Reshape = reshape
 | Invocation (fid, reshape)
;

end
 = Ok
 | Err
;

reshape
 =   Select {Path: path}
 |    ReMap [{Key: path, Value: path}]
 |    Set {Map: MapStrAny}
;



