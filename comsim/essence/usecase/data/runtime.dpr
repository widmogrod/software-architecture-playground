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
 |        End {Reason: endT}
 |     Choose {if: predicate, then: Activity, else: Activity}
 |    Reshape (reshapeT)
 | Invocation (fid)
;

endT
 = Ok
 | Err
;

reshapeT
 = rpath [lit]
 | rlist [path]
 | rdict [(path, reshapeT)]
;

