predicate
 =     Eq {left: reshaper, right: reshaper}
 | Exists {path: path}
 |    And (predicate, predicate)
 |     Or (predicate, predicate)
;

workflow
 =   Activity {id: AID, activity: activityT}
 | Transition {from: workflow, to: workflow}
;

activityT
 =      Start {Var: string}
 |        End = end
 |     Choose {if: predicate, then: workflow, else: workflow}
 |     Assign {Var: string, flow: workflow}
 |    reshaper = reshaper
 | Invocation (fid, reshaper)
;

end
 =  Ok (reshaper)
 | Err (reshaper)
;

reshaper
 =  GetValue (path)
 |  SetValue (values)
;


values
  = VFloat (float64)
  | VInt (int64)
  | VString (string)
  | VBool (bool)
  | VMap [{Key: reshaper, Value: reshaper}]
  | VList [reshaper]
;

